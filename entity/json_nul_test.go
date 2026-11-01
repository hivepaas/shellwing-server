package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/base"
)

// settings.data is JSONB, and Postgres refuses the \u0000 JSON writes for a NUL:
// the statement fails, and the transaction it is in with it. The NUL is dropped
// instead, as a task's args and output drop it, and what is read afterwards is
// what was stored.
func TestSettingDataDropsNULCharacters(t *testing.T) {
	s := &Setting{Type: base.SettingTypeKeyAuth}

	assert.NoError(t, s.SetData(&KeyAuth{KeyID: "A\x00K\x00IA"}))

	assert.NotContains(t, s.Data, `\u0000`)
	assert.Equal(t, int64(len(s.Data)), s.Size)
	assert.Equal(t, "AKIA", s.MustAsKeyAuth().KeyID)
}

// A backslash before "u0000" is text, not a NUL: kept as it is.
func TestSettingDataKeepsTheTextOfAnEscape(t *testing.T) {
	s := &Setting{Type: base.SettingTypeKeyAuth}
	data := &KeyAuth{KeyID: `\u0000`}

	assert.NoError(t, s.SetData(data))

	assert.Same(t, data, s.MustAsKeyAuth(), "nothing dropped: the data set is the data read")
	assert.Equal(t, `\u0000`, s.MustAsKeyAuth().KeyID)
}

// A task's output has its NULs dropped, and what the output parses to is read
// from what was stored.
func TestATaskOutputDropsTheNULs(t *testing.T) {
	task := &Task{}
	assert.NoError(t, task.SetOutput(map[string]any{"error": "boom: \x00bad\x00 header"}))
	assert.NotContains(t, task.Output, `\u0000`)
	assert.Contains(t, task.Output, `boom: bad header`)
}

// The text "\u0000" - a backslash, then u0000 - is no NUL, and is kept; so is
// any number, exactly; and a clean value is stored as it marshals.
func TestMarshalJSONBKeepsWhatIsNoNUL(t *testing.T) {
	b, clean, err := marshalJSONB(map[string]any{"path": `C:\u0000dir`, "ns": int64(9007199254740993)})
	assert.NoError(t, err)
	assert.True(t, clean)
	assert.Equal(t, `{"ns":9007199254740993,"path":"C:\\u0000dir"}`, string(b))

	b, clean, err = marshalJSONB(map[string]any{"ns": int64(9007199254740993), "msg": "a\x00b",
		"list": []any{"c\x00"}, "k\x00ey": 1})
	assert.NoError(t, err)
	assert.False(t, clean)
	assert.False(t, strings.Contains(string(b), `\u0000`))
	assert.Equal(t, `{"key":1,"list":["c"],"msg":"ab","ns":9007199254740993}`, string(b))
}
