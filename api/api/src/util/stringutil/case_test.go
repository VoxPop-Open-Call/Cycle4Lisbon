package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamelToSnake(t *testing.T) {
	for i, tc := range []struct {
		val, exp string
	}{
		{"", ""},
		{"id", "id"},
		{"ID", "id"},
		{"name", "name"},
		{"Name", "name"},
		{"RegX", "reg_x"},
		{"RegExp", "reg_exp"},
		{"JSONData", "json_data"},
		{"mixedCaps", "mixed_caps"},
		{"MixedCaps", "mixed_caps"},
		{"createdAt", "created_at"},
		{"DOMObject", "dom_object"},
		{"articleURL", "article_url"},
		{"XMLEncoding", "xml_encoding"},
		{"XMLHttpRequest", "xml_http_request"},
		{"XMLEncodingUTF", "xml_encoding_utf"},
		{"XMLEncodingUTF8", "xml_encoding_utf8"},
	} {
		assert.Equal(t, tc.exp, CamelToSnake(tc.val),
			"failed on test %d: %s", i, tc.val)
	}
}
