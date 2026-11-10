package csvfile

import (
	"fmt"
	"net/mail"
	"unicode"
)

// maxEmailLength is the longest accepted address (RFC 5321 path limit).
const maxEmailLength = 254

// validateEmail accepts exactly one plain address ("a@b.de"): no display
// name, no list ("a@b.de; c@d.de"), no spaces or control characters. The
// value is written to ChurchTools and receives the account activation, so
// it is checked before anything is sent.
func validateEmail(value string) error {
	if len(value) > maxEmailLength {
		return fmt.Errorf("E-Mail-Adresse länger als %d Zeichen", maxEmailLength)
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return fmt.Errorf("E-Mail-Adresse %q enthält Leer- oder Steuerzeichen", value)
		}
	}
	addr, err := mail.ParseAddress(value)
	if err != nil || addr.Name != "" || addr.Address != value {
		return fmt.Errorf("ungültige E-Mail-Adresse %q (genau eine Adresse ohne Namen erwartet)", value)
	}
	return nil
}
