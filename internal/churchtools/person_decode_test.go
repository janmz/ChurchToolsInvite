package churchtools

import "testing"

func TestDecodePersonInvitationStatus(t *testing.T) {
	person, err := decodePerson([]byte(`{
		"id": 42,
		"invitationStatus": "accepted"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if !person.HasChurchToolsAccount() {
		t.Fatal("expected account from invitationStatus accepted")
	}
	if person.AccountStatusLabel() != "Einladung bereits angenommen" {
		t.Fatalf("label = %q", person.AccountStatusLabel())
	}
}

func TestDecodePersonIsSystemUserAsInt(t *testing.T) {
	person, err := decodePerson([]byte(`{
		"id": 42,
		"firstName": "Max",
		"lastName": "Muster",
		"email": "max@example.org",
		"isSystemUser": 1
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if !person.HasChurchToolsAccount() {
		t.Fatal("expected system user from numeric isSystemUser")
	}
}

func TestDecodePersonAcceptedSecurityCamelCase(t *testing.T) {
	person, err := decodePerson([]byte(`{
		"id": 42,
		"acceptedSecurity": "2024-01-15"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if !person.HasChurchToolsAccount() {
		t.Fatal("expected account from acceptedSecurity")
	}
}

func TestDecodePersonLastLogin(t *testing.T) {
	person, err := decodePerson([]byte(`{
		"id": 42,
		"lastLogin": "2024-01-15T10:00:00Z"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if !person.HasChurchToolsAccount() {
		t.Fatal("expected account from lastLogin")
	}
}
