package userdto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/pkg/translation"
)

// next ends up in a link the server mails under its own name, so anything a browser could take to another
// site is refused, including what only a browser's leniency turns into one.
func TestIsSafeNextPath(t *testing.T) {
	safe := []string{
		"/",
		"/account/licenses",
		"/account/licenses/",
		"/account/trial?installationId=q9xk2m4j7n1p0r5s8t3v6w2yz",
		"/pricing?plan=hive&period=yearly#compare",
		"/account/checkout?return=https://example.com", // a query is the web's to read, not a host
		"/a/%2e%2e/b",
		"/" + strings.Repeat("a", nextPathMaxLen-1),
	}
	for _, next := range safe {
		assert.True(t, IsSafeNextPath(next), "%q is a path on the web app", next)
	}

	unsafe := []string{
		"",
		"account/licenses", // relative to wherever the link is opened
		"https://evil.example.com/",
		"http:/evil.example.com",
		"javascript:alert(1)",
		"//evil.example.com",
		"///evil.example.com",
		`/\evil.example.com`,
		`/\/evil.example.com`,
		"/\t/evil.example.com", // browsers drop the tab
		"/\n/evil.example.com",
		"/ /evil.example.com",
		"/account//licenses",
		"/../evil",
		"/account/./licenses",
		"/account/..",
		"/ünïcode",
		"/" + strings.Repeat("a", nextPathMaxLen),
	}
	for _, next := range unsafe {
		assert.False(t, IsSafeNextPath(next), "%q must be refused", next)
	}
}

func TestRequestUserSignupReq(t *testing.T) {
	newReq := func() *RequestUserSignupReq {
		return &RequestUserSignupReq{
			Email:    "  New.Person@Example.COM ",
			FullName: "  New Person ",
			Next:     " /account/trial ",
		}
	}

	t.Run("normalized", func(t *testing.T) {
		req := newReq()
		require.NoError(t, req.ModifyRequest())
		assert.Equal(t, "new.person@example.com", req.Email)
		assert.Equal(t, "New Person", req.FullName)
		assert.Equal(t, "/account/trial", req.Next)
		assert.Empty(t, req.Validate())
	})

	t.Run("next is optional", func(t *testing.T) {
		req := newReq()
		req.Next = ""
		require.NoError(t, req.ModifyRequest())
		assert.Empty(t, req.Validate())
	})

	tests := []struct {
		name  string
		edit  func(req *RequestUserSignupReq)
		field string
	}{
		{"no email", func(req *RequestUserSignupReq) { req.Email = "" }, "email"},
		{"not an email", func(req *RequestUserSignupReq) { req.Email = "not-an-email" }, "email"},
		{"no name", func(req *RequestUserSignupReq) { req.FullName = " " }, "fullName"},
		{"name too long", func(req *RequestUserSignupReq) { req.FullName = strings.Repeat("a", nameMaxLen+1) },
			"fullName"},
		{"next on another site", func(req *RequestUserSignupReq) { req.Next = "https://evil.example.com" }, "next"},
		{"next protocol-relative", func(req *RequestUserSignupReq) { req.Next = "//evil.example.com" }, "next"},
		{"next too long", func(req *RequestUserSignupReq) { req.Next = "/" + strings.Repeat("a", nextPathMaxLen) },
			"next"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newReq()
			tt.edit(req)
			require.NoError(t, req.ModifyRequest())
			errs := req.Validate()
			require.NotEmpty(t, errs)
			var paths []string
			for _, inner := range errs.Build(translation.LangEn).InnerErrors {
				paths = append(paths, inner.Path)
			}
			assert.Contains(t, paths, tt.field)
			assert.NotContains(t, errs.Error(), "ERR_", "every message is translated")
		})
	}
}
