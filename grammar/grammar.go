package grammar

import "unicode"

// TitleCase returns a title-cased string
func TitleCase(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
