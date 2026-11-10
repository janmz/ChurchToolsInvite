package churchtools

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// ErrSingleEmailOnly is returned when the instance rejects the "emails" list
// and the single-address fallback was not allowed.
var ErrSingleEmailOnly = errors.New("ChurchTools lehnt die Adressliste ab; nur mit --allow-single-email wird die primäre Adresse ersetzt (die bisherige geht dabei verloren)")

// UpdatePersonEmail applies an e-mail sync plan via PATCH /persons/{id}.
// When the instance rejects the "emails" list as invalid (HTTP 400/422) and
// allowSingle is set, only the single "email" field is sent; that replaces
// the previous primary address, which legacy reports as true. Without
// allowSingle ErrSingleEmailOnly is returned and nothing is replaced.
func (c *Client) UpdatePersonEmail(personID int, plan EmailSyncPlan, allowSingle bool) (legacy bool, err error) {
	if !plan.Needed {
		return false, nil
	}

	payload := map[string]any{
		"email": plan.Primary,
	}
	if len(plan.Emails) > 0 {
		emails := make([]map[string]any, len(plan.Emails))
		for i, entry := range plan.Emails {
			item := map[string]any{
				"email":     entry.Email,
				"isDefault": entry.IsDefault,
			}
			if entry.ContactLabelID != 0 {
				item["contactLabelId"] = entry.ContactLabelID
			}
			emails[i] = item
		}
		payload["emails"] = emails
	}

	err = c.patchPerson(personID, payload)
	if err == nil {
		return false, nil
	}
	if len(plan.Emails) <= 1 || !isInvalidRequest(err) {
		return false, err
	}
	if !allowSingle {
		return false, fmt.Errorf("%w: %v", ErrSingleEmailOnly, err)
	}

	// Fallback for instances that only accept the legacy single email field.
	if err := c.patchPerson(personID, map[string]any{"email": plan.Primary}); err != nil {
		return false, fmt.Errorf(
			"e-mail aktualisieren fehlgeschlagen (mehrfach-adressen nicht unterstützt?): %w",
			err,
		)
	}
	return true, nil
}

func (c *Client) patchPerson(personID int, payload map[string]any) error {
	path := "/persons/" + strconv.Itoa(personID)
	status, body, err := c.callAPI(http.MethodPatch, path, nil, payload)
	if err != nil {
		return fmt.Errorf("person aktualisieren: %w", err)
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return newAPIError(status, "Person konnte nicht aktualisiert werden (Berechtigung Personen bearbeiten?)", body)
	}
	return nil
}
