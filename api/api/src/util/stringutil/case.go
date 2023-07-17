package stringutil

import (
	"strings"
	"unicode"
)

// CamelToSnake converts a camelCase (or PascalCase) string into snake_case.
func CamelToSnake(src string) string {
	sb := strings.Builder{}
	var prev, next rune
	for i, curr := range src {
		if i+1 < len(src) {
			next = []rune(src)[i+1]
		} else {
			next = 0
		}

		if unicode.IsUpper(curr) {
			if unicode.IsLower(prev) ||
				(unicode.IsUpper(prev) && unicode.IsLower(next)) {
				sb.WriteRune('_')
			}

			sb.WriteRune(unicode.ToLower(curr))
		} else {
			sb.WriteRune(curr)
		}

		prev = curr
	}
	return sb.String()
}
