package licensesig

import (
	"strings"
	"unicode"
)

// MaxKeyTextLen bounds the text taken as a key. It must match license.MaxKeyTextLen in HivePaaS, which decodes
// nothing longer: a key is a few kilobytes (the ML-DSA-65 signature alone is 3309 bytes), so this leaves room for an
// Apiary key naming many installations.
const MaxKeyTextLen = 32 << 10

// NormalizeKeyText removes every whitespace character from s, as license.NormalizeKeyText in HivePaaS does. Nothing
// in an envelope holds any (its strings are base64, hex, key ids and algorithm names), so this changes nothing a key
// says, and a key wrapped by an email client or indented in a file is the same key.
func NormalizeKeyText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
