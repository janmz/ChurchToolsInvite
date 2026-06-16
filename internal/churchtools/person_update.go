package churchtools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// UpdatePersonEmail applies an e-mail sync plan via PATCH /persons/{id}.
func (c *Client) UpdatePersonEmail(personID int, plan EmailSyncPlan) error {
	if !plan.Needed {
		return nil
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

	if err := c.patchPerson(personID, payload); err == nil {
		return nil
	} else if len(plan.Emails) <= 1 {
		return err
	}

	// Fallback for instances that only accept the legacy single email field.
	if err := c.patchPerson(personID, map[string]any{"email": plan.Primary}); err != nil {
		return fmt.Errorf(
			"e-mail aktualisieren fehlgeschlagen (mehrfach-adressen nicht unterstützt?): %w",
			err,
		)
	}
	return nil
}

func (c *Client) patchPerson(personID int, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	path := "/persons/" + strconv.Itoa(personID)
	req, err := http.NewRequest(http.MethodPatch, c.apiURL(path), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.applyAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("person aktualisieren: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.relogin(); err != nil {
			return err
		}
		return c.patchPerson(personID, payload)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "person konnte nicht aktualisiert werden (Berechtigung Personen bearbeiten?)",
			Body:       string(respBody),
		}
	}
	return nil
}
