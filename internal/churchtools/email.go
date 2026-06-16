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

	if current == "" {
		labelID := defaultContactLabelID(existing)
		return EmailSyncPlan{
			Needed:  true,
			Primary: csvEmail,
			Emails: []PersonEmail{{
				Email:          csvEmail,
				IsDefault:      true,
				ContactLabelID: labelID,
			}},
			Detail: fmt.Sprintf("primäre e-mail gesetzt auf %s", csvEmail),
		}
	}

	if strings.EqualFold(current, csvEmail) {
		return EmailSyncPlan{Detail: "e-mail bereits identisch"}
	}

	labelForNewDefault := defaultContactLabelID(existing)
	oldLabelID := contactLabelForEmail(existing, current)
	if oldLabelID == 0 {
		oldLabelID = labelForNewDefault
	}

	updated := make([]PersonEmail, 0, len(existing)+2)
	csvPresent := false

	for _, entry := range existing {
		addr := strings.TrimSpace(entry.Email)
		if addr == "" {
			continue
		}
		if strings.EqualFold(addr, csvEmail) {
			csvPresent = true
			updated = append(updated, PersonEmail{
				Email:          addr,
				IsDefault:      true,
				ContactLabelID: entry.ContactLabelID,
			})
			continue
		}
		updated = append(updated, PersonEmail{
			Email:          addr,
			IsDefault:      false,
			ContactLabelID: entry.ContactLabelID,
		})
	}

	if !csvPresent {
		updated = append([]PersonEmail{{
			Email:          csvEmail,
			IsDefault:      true,
			ContactLabelID: labelForNewDefault,
		}}, updated...)
	}

	if !containsEmail(updated, current) {
		updated = append(updated, PersonEmail{
			Email:          current,
			IsDefault:      false,
			ContactLabelID: oldLabelID,
		})
	}

	return EmailSyncPlan{
		Needed:  true,
		Primary: csvEmail,
		Emails:  updated,
		Detail: fmt.Sprintf(
			"primäre e-mail %s -> %s, bisherige als zusätzliche behalten: %s",
			current,
			csvEmail,
			current,
		),
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

func containsEmail(entries []PersonEmail, email string) bool {
	for _, entry := range entries {
		if strings.EqualFold(strings.TrimSpace(entry.Email), email) {
			return true
		}
	}
	return false
}
