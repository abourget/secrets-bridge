package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecrets(t *testing.T) {
	funkyBytes := []byte{124, 125, 126, 127, 128, 129, 174, 175, 176, 177, 181, 182, 189, 190, 191, 192}

	tests := []struct {
		note        string
		addKey      string
		addValue    []byte
		addSucceeds bool
		getKey      string
		expectValue []byte
	}{
		{
			"Happy path",
			"hello", b("world"), true,
			"hello", b("world"),
		},
		{
			"Encoded happy path",
			"b64:key", b("fH1+f4CBrq+wsbW2vb6/wA=="), true,
			"key", funkyBytes,
		},
		{
			"Error path",
			"b64:key", b("invalid base64"), false,
			"", nil,
		},
		{
			"Key does not exist",
			"key1", b("val"), true,
			"key2", nil,
		},
		{
			"b64 decode source value",
			"b64:hello", b("d29ybGQ="), true,
			"hello", b("world"),
		},
		{
			"b64 encoding chars",
			"key", funkyBytes, true,
			"b64:key", b("fH1+f4CBrq+wsbW2vb6/wA=="),
		},
		{
			"b64u encoding chars",
			"key", funkyBytes, true,
			"b64u:key", b("fH1-f4CBrq-wsbW2vb6_wA=="),
		},
		{
			"rb64 encoding chars",
			"key", funkyBytes, true,
			"rb64:key", b("fH1+f4CBrq+wsbW2vb6/wA"),
		},
		{
			"rb64u encoding chars",
			"key", funkyBytes, true,
			"rb64u:key", b("fH1-f4CBrq-wsbW2vb6_wA"),
		},
		{
			"invalid b64 decoding chars",
			"b64:key", b("fH1-f4CBrq-wsbW2vb6_wA=="), false,  // invalid chars for specified encoding...
			"key", nil,
		},
		{
			"b64 decoding chars",
			"b64:key", b("fH1+f4CBrq+wsbW2vb6/wA=="), true,
			"key", funkyBytes,
		},
		{
			"b64u decoding chars",
			"b64u:key", b("fH1-f4CBrq-wsbW2vb6_wA=="), true,
			"key", funkyBytes,
		},
		{
			"rb64 decoding chars",
			"rb64:key", b("fH1+f4CBrq+wsbW2vb6/wA"), true,
			"key", funkyBytes,
		},
		{
			"rb64u decoding chars",
			"rb64u:key", b("fH1-f4CBrq-wsbW2vb6_wA"), true,
			"key", funkyBytes,
		},
	}

	for _, test := range tests {
		store := &Store{}
		err := store.Add(test.addKey, test.addValue)
		if test.addSucceeds {
			assert.NoError(t, err, "Test: "+test.note)

			getVal := store.Get(test.getKey)
			assert.Equal(t, test.expectValue, getVal, "Test: "+test.note)

		} else {
			assert.Error(t, err, "Test: "+test.note)
		}
	}
}

func b(s string) []byte {
	return []byte(s)
}
