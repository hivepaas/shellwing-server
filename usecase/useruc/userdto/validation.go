package userdto

import (
	"path/filepath"
	"strings"

	"github.com/asaskevich/govalidator"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
)

const (
	nameMinLen = 1
	nameMaxLen = 100

	notesMinLen = 1
	notesMaxLen = 10000
)

func validateUsername(username *string, required bool, field string) (res []vld.Validator) {
	res = append(res, basedto.ValidateStr(username, required,
		nameMinLen, nameMaxLen, "username")...)

	// NOTE: username must not be a valid email address as it takes the address
	if username != nil {
		res = append(res, vld.Must(!govalidator.IsEmail(*username)).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_ARGUMENT_INVALID"),
		))
	}
	return res
}

// validateAdminSecurityOption refuses an admin account that authenticates with a
// password and nothing else.
//
// Usable on the invite request alone because both halves are required there. The
// update request carries them as optional pointers, so the same rule has to be
// applied against the merged result in the usecase - see ensureAdminSecurityOption.
func validateAdminSecurityOption(
	role base.UserRole,
	option base.UserSecurityOption,
	field string,
) []vld.Validator {
	return []vld.Validator{
		vld.Must(base.SecurityOptionAllowedForRole(role, option)).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_ADMIN_SECURITY_OPTION_WEAK"),
		),
	}
}

func validateUserPhoto(photo *UserPhotoReq, field string) (res []vld.Validator) {
	if photo == nil || photo.FileName == "" {
		return nil
	}
	fileExt := strings.ToLower(filepath.Ext(photo.FileName))
	return []vld.Validator{
		vld.Must(gofn.Contain(base.AllPhotoFileExts, fileExt)).OnError(
			vld.SetField(field+".fileName", nil),
			vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_EXT_UNSUPPORTED"),
		),
		vld.Must(len(photo.DataBytes) > 0 && base.IsValidPhotoContent(photo.DataBytes, fileExt)).OnError(
			vld.SetField(field+".dataBase64", nil),
			vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_INVALID"),
		),
		vld.When(len(photo.DataBytes) > 0).Then(
			vld.Must(len(photo.DataBytes) <= base.PhotoMaxSize).OnError(
				vld.SetField(field+".dataBase64", nil),
				vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_TOO_BIG"),
			),
		),
	}
}
