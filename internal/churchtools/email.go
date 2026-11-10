package churchtools

import (
	"fmt"
	"strings"
)

// PersonEmail is one ChurchTools e-mail entry with label metadata.
type PersonEmail struct {
	Email          string `json:"email"`
	IsDefault      bool   `json:"isDefault"`
	ContactLabelID int    `json:"contactLabelId"`
}

// EmailSyncPlan describes a person e-mail update before inviting.
type EmailSyncPlan struct {
	Needed  bool
	Primary string
	Emails  []PersonEmail
	Detail  string
}

// PrepareEmailUpdate builds a sync plan from CSV and ChurchTools person data.
// The CSV e-mail becomes primary; the previous primary is kept as additional
// when possible.
func PrepareEmailUpdate(csvEmail string, person Person) EmailSyncPlan {
	csvEmail = strings.TrimSpace(csvEmail)
	if csvEmail == "" {
		return EmailSyncPlan{}
	}

	current := strings.TrimSpace(person.Email)
	existing := normalizePersonEmails(person.Emails, current)

	if current != "" && strings.EqualFold(current, csvEmail) && isDefaultEmail(existing, csvEmail) {
		return EmailSyncPlan{Detail: "e-mail bereits identisch"}
	}

	if current != "" && strings.EqualFold(current, csvEmail) {
		return EmailSyncPlan{Detail: "e-mail bereits identisch"}
	}

	labelForNewDefault := defaultContactLabelID(existing)
	oldLabelID := contactLabelForEmail(existing, current)
	if oldLabelID == 0 {
		oldLabelID = labelForNewDefault
	}

	updated := make([]PersonEmail, 0, len(existing)+2)
	seen := make(map[string]int, len(existing))
	csvKey := strings.ToLower(csvEmail)

	for _, entry := range existing {
		addr := strings.TrimSpace(entry.Email)
		if addr == "" {
			continue
		}
		key := strings.ToLower(addr)
		if idx, ok := seen[key]; ok {
			// De-duplicate: keep a single entry per address.
			// If the address matches the CSV e-mail, ensure it becomes default.
			if key == csvKey && !updated[idx].IsDefault {
				updated[idx].IsDefault = true
			}
			continue
		}
		seen[key] = len(updated)
		updated = append(updated, PersonEmail{
			Email:          addr,
			IsDefault:      key == csvKey,
			ContactLabelID: entry.ContactLabelID,
		})
	}

	if _, ok := seen[csvKey]; !ok {
		updated = append([]PersonEmail{{
			Email:          csvEmail,
			IsDefault:      true,
			ContactLabelID: labelForNewDefault,
		}}, updated...)
		// Shift indexes of previously seen entries by 1 (we prepended).
		for key, idx := range seen {
			seen[key] = idx + 1
		}
		seen[csvKey] = 0
	}

	currentKey := strings.ToLower(strings.TrimSpace(current))
	if currentKey != "" {
		if idx, ok := seen[currentKey]; ok {
			if updated[idx].IsDefault {
				updated[idx].IsDefault = false
			}
		} else {
			updated = append(updated, PersonEmail{
				Email:          current,
				IsDefault:      false,
				ContactLabelID: oldLabelID,
			})
		}
	}

	// Without a primary address ("email" empty) the other entries in
	// "emails" are kept as well: the PATCH replaces the whole list.
	detail := fmt.Sprintf("primäre e-mail gesetzt auf %s", csvEmail)
	if current != "" {
		detail = fmt.Sprintf("primäre e-mail %s -> %s, bisherige als zusätzliche behalten: %s", current, csvEmail, current)
	}
	return EmailSyncPlan{
		Needed:  true,
		Primary: csvEmail,
		Emails:  updated,
		Detail:  detail,
	}
}

func normalizePersonEmails(entries []PersonEmail, primary string) []PersonEmail {
	if len(entries) > 0 {
		return entries
	}
	primary = strings.TrimSpace(primary)
	if primary == "" {
		return nil
	}
	return []PersonEmail{{
		Email:          primary,
		IsDefault:      true,
		ContactLabelID: 0,
	}}
}

func isDefaultEmail(entries []PersonEmail, email string) bool {
	if len(entries) == 0 {
		return true
	}
	for _, entry := range entries {
		if strings.EqualFold(strings.TrimSpace(entry.Email), email) {
			return entry.IsDefault
		}
	}
	return false
}

func defaultContactLabelID(entries []PersonEmail) int {
	for _, entry := range entries {
		if entry.IsDefault && entry.ContactLabelID != 0 {
			return entry.ContactLabelID
		}
	}
	for _, entry := range entries {
		if entry.ContactLabelID != 0 {
			return entry.ContactLabelID
		}
	}
	return 1
}

func contactLabelForEmail(entries []PersonEmail, email string) int {
	for _, entry := range entries {
		if strings.EqualFold(strings.TrimSpace(entry.Email), email) {
			return entry.ContactLabelID
		}
	}
	return 0
}
