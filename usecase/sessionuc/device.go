package sessionuc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/entity/cacheentity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

const (
	deviceSecretLen   = 32
	deviceCallback    = "/callback"
	deviceAuthCodeTTL = 2 * time.Minute // the app is waiting
)

// How an app signs in (RFC 8252): it opens the browser at the server, whatever sign-in the server has happens there,
// and the person approves the device. The server answers with a code, which goes to the app's loopback address, and
// the app exchanges it for tokens with the verifier only it has (RFC 7636).
//
// Its tokens reach sync and nothing else: an admin's device cannot manage the server.

func randomSecret() string {
	data := make([]byte, deviceSecretLen)
	_, _ = rand.Read(data) // never fails: see crypto/rand
	return hex.EncodeToString(data)
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// validDeviceRedirect says whether the address is an app's loopback listener, the only place a code may go. Anything
// else would let a web page have the browser hand it the code. Literal addresses, as RFC 8252 asks: a name could
// resolve elsewhere.
func validDeviceRedirect(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "http" || parsed.Path != deviceCallback || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	ip := net.ParseIP(parsed.Hostname())
	if ip == nil || !ip.IsLoopback() {
		return false
	}
	port, err := strconv.Atoi(parsed.Port())
	return err == nil && port > 0
}

func pkceMatches(verifier, challenge string) bool {
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

// AuthorizeDevice is the person's approval, from their browser's session.
func (uc *UC) AuthorizeDevice(ctx context.Context, auth *basedto.Auth,
	req *sessiondto.AuthorizeDeviceReq) (*sessiondto.AuthorizeDeviceResp, error) {
	if !validDeviceRedirect(req.RedirectURI) {
		return nil, hperrors.Wrap(hperrors.ErrDeviceRedirectInvalid)
	}
	code := randomSecret()
	err := uc.deviceAuthCodeRepo.Set(ctx, hashSecret(code), &cacheentity.DeviceAuthCode{
		UserID:        auth.UserID(),
		CodeChallenge: req.CodeChallenge,
		RedirectURI:   req.RedirectURI,
		DeviceName:    req.DeviceName,
		Platform:      req.Platform,
	}, deviceAuthCodeTTL)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	query := url.Values{"code": {code}, "state": {req.State}}
	return &sessiondto.AuthorizeDeviceResp{
		Data: &sessiondto.AuthorizeDeviceData{RedirectURL: req.RedirectURI + "?" + query.Encode()},
	}, nil
}

// ExchangeDeviceCode makes the device and gives it its tokens. A code works once, whatever becomes of the attempt.
func (uc *UC) ExchangeDeviceCode(ctx context.Context,
	req *sessiondto.ExchangeDeviceCodeReq) (*sessiondto.DeviceTokensResp, error) {
	code, err := uc.deviceAuthCodeRepo.Take(ctx, hashSecret(req.Code))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if code == nil || code.RedirectURI != req.RedirectURI || !pkceMatches(req.CodeVerifier, code.CodeChallenge) {
		return nil, hperrors.Wrap(hperrors.ErrDeviceCodeInvalid)
	}
	user, err := uc.userRepo.GetByID(ctx, uc.db, code.UserID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = signInRefusal(user); err != nil {
		return nil, err
	}

	timeNow := timeutil.NowUTC()
	refreshToken := randomSecret()
	device := &entity.Device{
		ID:               gofn.Must(ulid.NewStringULID()),
		UserID:           user.ID,
		Name:             code.DeviceName,
		Platform:         code.Platform,
		RefreshTokenHash: hashSecret(refreshToken),
		CreatedAt:        timeNow,
		UpdatedAt:        timeNow,
		LastSeenAt:       timeNow,
		RefreshExpireAt:  timeNow.Add(config.Current().Session.DeviceRefreshTokenExp),
	}
	if err = uc.deviceRepo.Insert(ctx, uc.db, device); err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := uc.issueDeviceAccess(ctx, device)
	if err != nil {
		return nil, err
	}
	data.RefreshToken = refreshToken
	if err = uc.recordLogin(ctx, user, auditMethodDevice); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &sessiondto.DeviceTokensResp{Data: data}, nil
}

// RefreshDevice gives a device a new access token. The refresh token stays as it is, its life starting over: a device
// that lost an answer to a rotation would be signed out for good, and it is stored encrypted in the vault.
func (uc *UC) RefreshDevice(ctx context.Context,
	req *sessiondto.DeviceRefreshTokenReq) (*sessiondto.DeviceTokensResp, error) {
	timeNow := timeutil.NowUTC()
	device, err := uc.deviceRepo.GetByRefreshTokenHash(ctx, uc.db, hashSecret(req.RefreshToken))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if device == nil || !device.CanRefresh(timeNow) {
		return nil, hperrors.Wrap(hperrors.ErrDeviceTokenInvalid)
	}
	user, err := uc.userRepo.GetByID(ctx, uc.db, device.UserID)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(hperrors.ErrDeviceTokenInvalid).WithCause(err)
		}
		return nil, hperrors.Wrap(err)
	}
	if err = signInRefusal(user); err != nil {
		return nil, err
	}
	device.LastSeenAt = timeNow
	device.UpdatedAt = timeNow
	device.RefreshExpireAt = timeNow.Add(config.Current().Session.DeviceRefreshTokenExp)
	if err = uc.deviceRepo.Update(ctx, uc.db, device); err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := uc.issueDeviceAccess(ctx, device)
	if err != nil {
		return nil, err
	}
	return &sessiondto.DeviceTokensResp{Data: data}, nil
}

// issueDeviceAccess makes an access token whose session is the device itself, so that revoking the device ends it.
func (uc *UC) issueDeviceAccess(ctx context.Context, device *entity.Device) (*sessiondto.DeviceTokensData, error) {
	exp := config.Current().Session.DeviceAccessTokenExp
	claims := &jwtsession.AuthClaims{
		UID:       device.ID,
		UserID:    device.UserID,
		Scope:     jwtsession.ScopeSync,
		StartedAt: timeutil.NowUTC().Unix(),
	}
	token, err := jwtsession.GenerateAccessTokenWithExp(claims, exp)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create access token")
	}
	if err = uc.userTokenRepo.Set(ctx, device.UserID, device.ID, exp); err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to store token in cache")
	}
	return &sessiondto.DeviceTokensData{
		DeviceID:       device.ID,
		AccessToken:    token,
		AccessTokenExp: claims.ExpiresAt.Time,
	}, nil
}

// SignOutDevice is the app signing out, with its own refresh token. A token no device has is no error: the app's
// aim is that the token no longer works.
func (uc *UC) SignOutDevice(ctx context.Context,
	req *sessiondto.DeviceRefreshTokenReq) (*sessiondto.SignOutDeviceResp, error) {
	device, err := uc.deviceRepo.GetByRefreshTokenHash(ctx, uc.db, hashSecret(req.RefreshToken))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if device != nil {
		if err = uc.revokeDevice(ctx, device); err != nil {
			return nil, err
		}
	}
	return &sessiondto.SignOutDeviceResp{}, nil
}

func (uc *UC) ListDevices(ctx context.Context, auth *basedto.Auth) (*sessiondto.ListDevicesResp, error) {
	devices, err := uc.deviceRepo.ListOfUser(ctx, uc.db, auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &sessiondto.ListDevicesResp{Data: sessiondto.TransformDevices(devices)}, nil
}

// RevokeDevice signs out one of the person's devices. Another person's is simply not found.
func (uc *UC) RevokeDevice(ctx context.Context, auth *basedto.Auth,
	deviceID string) (*sessiondto.RevokeDeviceResp, error) {
	device, err := uc.deviceRepo.GetOfUser(ctx, uc.db, auth.UserID(), deviceID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if device == nil {
		return nil, hperrors.NewNotFound("Device")
	}
	if err = uc.revokeDevice(ctx, device); err != nil {
		return nil, err
	}
	return &sessiondto.RevokeDeviceResp{}, nil
}

func (uc *UC) revokeDevice(ctx context.Context, device *entity.Device) error {
	timeNow := timeutil.NowUTC()
	device.RevokedAt = timeNow
	device.UpdatedAt = timeNow
	device.RefreshTokenHash = ""
	if err := uc.deviceRepo.Update(ctx, uc.db, device); err != nil {
		return hperrors.Wrap(err)
	}
	// Its access token stops now, not when it expires.
	if err := uc.userTokenRepo.Del(ctx, device.UserID, device.ID); err != nil {
		return hperrors.Wrap(err)
	}
	// So do the SSH certificates it holds, in every organization. Signing a device out otherwise leaves whoever has
	// the laptop able to log in to servers until each certificate ends, which is the case this is for.
	_, err := uc.sshCertRepo.RevokeOfDevice(ctx, uc.db, device.ID, timeNow)
	return hperrors.Wrap(err)
}
