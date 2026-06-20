package churchtools

import (
	"encoding/json"
	"strings"
)

// PrivacyPolicyAgreement holds consent metadata for a person record.
type PrivacyPolicyAgreement struct {
	Date *string `json:"date"`
}

// UnmarshalJSON accepts object or array payloads from ChurchTools.
func (p *PrivacyPolicyAgreement) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	type plain PrivacyPolicyAgreement

	if data[0] == '[' {
		var items []plain
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		for _, item := range items {
			agreement := PrivacyPolicyAgreement(item)
			if hasStringValue(agreement.Date) {
				*p = agreement
				return nil
			}
		}
		return nil
	}

	var item plain
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	*p = PrivacyPolicyAgreement(item)
	return nil
}

// HasChurchToolsAccount reports whether the person already has or was invited to
// a ChurchTools user account.
func (p Person) HasChurchToolsAccount() bool {
	switch normalizeInvitationStatus(p.InvitationStatus) {
	case "accepted", "pending":
		return true
	}
	if p.IsSystemUser != nil && *p.IsSystemUser {
		return true
	}
	if p.IsAllowedToLogin != nil && *p.IsAllowedToLogin {
		return true
	}
	if strings.TrimSpace(p.CMSUserID) != "" {
		return true
	}
	if hasStringValue(p.AcceptedSecurity) {
		return true
	}
	if hasStringValue(p.LastLogin) {
		return true
	}
	if p.PrivacyPolicyAgreement != nil && hasStringValue(p.PrivacyPolicyAgreement.Date) {
		return true
	}
	return false
}

// IsRegisteredUser reports whether the person already has an active ChurchTools
// account (export status "Registriert"). Such persons must never be re-invited.
func (p Person) IsRegisteredUser() bool {
	return p.ExportStatusLabel() == "Registriert"
}

// ExportStatusLabel is the invitation status written to export CSV rows.
func (p Person) ExportStatusLabel() string {
	switch normalizeInvitationStatus(p.InvitationStatus) {
	case "pending":
		return "Eingeladen"
	case "accepted":
		return "Registriert"
	}
	if p.HasChurchToolsAccount() {
		return "Registriert"
	}
	return "NEU"
}

// AccountStatusLabel describes why a person is treated as already invited.
func (p Person) AccountStatusLabel() string {
	switch normalizeInvitationStatus(p.InvitationStatus) {
	case "accepted":
		return "Einladung bereits angenommen"
	case "pending":
		return "Einladung bereits versendet"
	}
	switch {
	case p.IsSystemUser != nil && *p.IsSystemUser:
		return "bereits ChurchTools-Benutzer"
	case p.IsAllowedToLogin != nil && *p.IsAllowedToLogin:
		return "darf sich bereits anmelden"
	case strings.TrimSpace(p.CMSUserID) != "":
		return "bereits Benutzername vergeben"
	case hasStringValue(p.AcceptedSecurity):
		return "Vertraulichkeitsvereinbarung bereits akzeptiert"
	case hasStringValue(p.LastLogin):
		return "bereits eingeloggt"
	case p.PrivacyPolicyAgreement != nil && hasStringValue(p.PrivacyPolicyAgreement.Date):
		return "Datenschutz-Einwilligung bereits erteilt"
	default:
		return "bereits eingeladen"
	}
}

func hasStringValue(value *string) bool {
	return value != nil && strings.TrimSpace(*value) != ""
}

func normalizeInvitationStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}
