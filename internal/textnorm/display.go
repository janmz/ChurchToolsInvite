package textnorm

import (
	"strings"
	"unicode"
)

// Printable makes text from ChurchTools or a CSV file safe for terminal
// output: control characters (incl. the ESC of ANSI sequences) and
// bidirectional overrides, which can hide or reorder output, become spaces.
func Printable(text string) string {
	if !strings.ContainsFunc(text, unsafeRune) {
		return text
	}
	return strings.Map(func(r rune) rune {
		if unsafeRune(r) {
			return ' '
		}
		return r
	}, text)
}

func unsafeRune(r rune) bool {
	if r == '\n' {
		return false
	}
	if unicode.IsControl(r) {
		return true
	}
	switch {
	case r >= 0x202A && r <= 0x202E, r >= 0x2066 && r <= 0x2069, r == 0x200E, r == 0x200F:
		return true
	}
	return false
}
