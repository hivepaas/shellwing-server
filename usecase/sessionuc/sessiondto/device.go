package sessiondto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
)

const (
	devicePlatformMaxLen = 50
	deviceNameMaxLen     = 100
	// RFC 7636: a verifier is 43 to 128 characters, and S256 of it 43 in base64url.
	pkceVerifierMinLen = 43
	pkceVerifierMaxLen = 128
	pkceChallengeLen   = 43
	deviceURLMaxLen    = 100
	deviceStateMaxLen  = 200
	deviceTokenMaxLen  = 200
)

// AuthorizeDeviceReq is a signed-in person approving an app on their device.
type AuthorizeDeviceReq struct {
	RedirectURI   string `json:"redirectUri"`
	CodeChallenge string `json:"codeChallenge"`
	State         string `json:"state"`
	DeviceName    string `json:"deviceName"`
	Platform      string `json:"platform"`
}

func NewAuthorizeDeviceReq() *AuthorizeDeviceReq {
	return &AuthorizeDeviceReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *AuthorizeDeviceReq) Validate() hperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.RedirectURI, true, 1, deviceURLMaxLen, "redirectUri")...)
	validators = append(validators, basedto.ValidateStr(&req.CodeChallenge, true, pkceChallengeLen, pkceChallengeLen,
		"codeChallenge")...)
	validators = append(validators, basedto.ValidateStr(&req.State, true, 1, deviceStateMaxLen, "state")...)
	validators = append(validators, basedto.ValidateStr(&req.DeviceName, true, 1, deviceNameMaxLen, "deviceName")...)
	validators = append(validators, basedto.ValidateStr(&req.Platform, false, 0, devicePlatformMaxLen, "platform")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type AuthorizeDeviceData struct {
	// RedirectURL is where the browser goes next: the app, with its code.
	RedirectURL string `json:"redirectUrl"`
}

type AuthorizeDeviceResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *AuthorizeDeviceData `json:"data"`
}

// ExchangeDeviceCodeReq is the app coming for its tokens.
type ExchangeDeviceCodeReq struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"codeVerifier"`
	RedirectURI  string `json:"redirectUri"`
}

func NewExchangeDeviceCodeReq() *ExchangeDeviceCodeReq {
	return &ExchangeDeviceCodeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ExchangeDeviceCodeReq) Validate() hperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.Code, true, 1, deviceTokenMaxLen, "code")...)
	validators = append(validators, basedto.ValidateStr(&req.CodeVerifier, true, pkceVerifierMinLen, pkceVerifierMaxLen,
		"codeVerifier")...)
	validators = append(validators, basedto.ValidateStr(&req.RedirectURI, true, 1, deviceURLMaxLen, "redirectUri")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// DeviceTokensData are a device's tokens. The refresh token comes once, with the code: a refresh answers without it.
type DeviceTokensData struct {
	DeviceID       string    `json:"deviceId"`
	AccessToken    string    `json:"accessToken"`
	AccessTokenExp time.Time `json:"accessTokenExp"`
	RefreshToken   string    `json:"refreshToken,omitempty"`
}

type DeviceTokensResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *DeviceTokensData `json:"data"`
}

// DeviceRefreshTokenReq names a device by its refresh token.
type DeviceRefreshTokenReq struct {
	RefreshToken string `json:"refreshToken"`
}

func NewDeviceRefreshTokenReq() *DeviceRefreshTokenReq {
	return &DeviceRefreshTokenReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeviceRefreshTokenReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateStr(&req.RefreshToken, true, 1, deviceTokenMaxLen, "refreshToken")
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type SignOutDeviceResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type DeviceResp struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Platform   string    `json:"platform"`
	CreatedAt  time.Time `json:"createdAt"`
	LastSeenAt time.Time `json:"lastSeenAt,omitzero"`
}

func TransformDevices(devices []*entity.Device) []*DeviceResp {
	resp := make([]*DeviceResp, 0, len(devices))
	for _, device := range devices {
		resp = append(resp, &DeviceResp{
			ID:         device.ID,
			Name:       device.Name,
			Platform:   device.Platform,
			CreatedAt:  device.CreatedAt,
			LastSeenAt: device.LastSeenAt,
		})
	}
	return resp
}

type ListDevicesResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*DeviceResp `json:"data"`
}

type RevokeDeviceResp struct {
	Meta *basedto.Meta `json:"meta"`
}
