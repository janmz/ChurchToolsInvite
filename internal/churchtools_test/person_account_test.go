package churchtools_test

import (
	"encoding/json"
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

func TestHasChurchToolsAccount(t *testing.T) {
	trueVal := true
	date := "2024-01-01"
	security := "2024-01-02"

	cases := []struct {
		name   string
		person churchtools.Person
		want   bool
	}{
		{
			name:   "empty",
			person: churchtools.Person{},
			want:   false,
		},
		{
			name:   "invitation accepted",
			person: churchtools.Person{InvitationStatus: "accepted"},
			want:   true,
		},
		{
			name:   "invitation pending",
			person: churchtools.Person{InvitationStatus: "pending"},
			want:   true,
		},
		{
			name:   "system user",
			person: churchtools.Person{IsSystemUser: &trueVal},
			want:   true,
		},
		{
			name:   "cms user id",
			person: churchtools.Person{CMSUserID: "max.muster"},
			want:   true,
		},
		{
			name:   "accepted security",
			person: churchtools.Person{AcceptedSecurity: &security},
			want:   true,
		},
		{
			name: "privacy policy",
			person: churchtools.Person{
				PrivacyPolicyAgreement: &churchtools.PrivacyPolicyAgreement{Date: &date},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.person.HasChurchToolsAccount(); got != tc.want {
				t.Fatalf("HasChurchToolsAccount() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPrivacyPolicyAgreementUnmarshalArray(t *testing.T) {
	var person churchtools.Person
	if err := json.Unmarshal([]byte(`{
		"id": 1,
		"privacyPolicyAgreement": [{"date": "2024-05-01"}]
	}`), &person); err != nil {
		t.Fatal(err)
	}
	if !person.HasChurchToolsAccount() {
		t.Fatal("expected account from array privacyPolicyAgreement")
	}
}

func TestPrivacyPolicyAgreementUnmarshalObject(t *testing.T) {
	var person churchtools.Person
	if err := json.Unmarshal([]byte(`{
		"id": 1,
		"privacyPolicyAgreement": {"date": "2024-05-01"}
	}`), &person); err != nil {
		t.Fatal(err)
	}
	if !person.HasChurchToolsAccount() {
		t.Fatal("expected account from object privacyPolicyAgreement")
	}
}
