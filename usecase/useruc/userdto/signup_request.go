package userdto

import (
	"net/url"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/strutil"
)

const (
	nextPathMaxLen = 500
)

type RequestUserSignupReq struct {
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	// Next is the page the web opens once the account is finished, such as /account/trial: a path on the
	// web app, carried through the emailed link. Optional.
	Next string `json:"next"`
}

func NewRequestUserSignupReq() *RequestUserSignupReq {
	return &RequestUserSignupReq{}
}

func (req *RequestUserSignupReq) ModifyRequest() error {
	req.Email = strutil.NormalizeEmail(req.Email)
	req.FullName = strings.TrimSpace(req.FullName)
	req.Next = strings.TrimSpace(req.Next)
	return nil
}

func (req *RequestUserSignupReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateEmail(&req.Email, true, "email")...)
	validators = append(validators, basedto.ValidateStr(&req.FullName, true,
		nameMinLen, nameMaxLen, "fullName")...)
	validators = append(validators, validateNextPath(&req.Next, "next")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type RequestUserSignupResp struct {
	Meta *basedto.Meta `json:"meta"`
}

func validateNextPath(next *string, field string) (res []vld.Validator) {
	res = append(res, basedto.ValidateStr(next, false, 1, nextPathMaxLen, field)...)
	if next != nil && *next != "" {
		res = append(res, vld.Must(IsSafeNextPath(*next)).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_URL_INVALID"),
		))
	}
	return res
}

// IsSafeNextPath reports whether next is a path on the web app's own origin, and nothing a browser could
// read as another site. The link it travels in is mailed by the server, so an open redirect here would
// send the server's own users wherever whoever asked for the mail chose.
//
// Browsers forgive more than url.Parse does: they drop tabs and newlines, read a backslash as a slash,
// and take "//host" or "/\host" as another host. So the text is held to printable ASCII without a
// backslash, to a single leading slash, and to no empty or dot segment that could collapse into "//".
func IsSafeNextPath(next string) bool {
	if len(next) == 0 || len(next) > nextPathMaxLen || next[0] != '/' {
		return false
	}
	for i := range len(next) {
		if c := next[i]; c <= ' ' || c >= 0x7f || c == '\\' {
			return false
		}
	}
	u, err := url.Parse(next)
	if err != nil || u.Scheme != "" || u.Host != "" || u.User != nil || u.Opaque != "" {
		return false
	}
	// The path proper, before any query or fragment: those may hold anything once encoded.
	path := next
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		path = path[:i]
	}
	for _, segment := range strings.Split(path, "/")[1:] {
		switch segment {
		case ".", "..":
			return false
		case "":
			// Only the trailing slash of a path may leave an empty segment: "/" or "/account/".
			if !strings.HasSuffix(path, "/") || strings.Contains(path, "//") {
				return false
			}
		}
	}
	return true
}
