package churchtools_test

import (
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

func TestPrepareEmailUpdateNoChange(t *testing.T) {
	person := churchtools.Person{
		Email: "same@example.org",
		Emails: []churchtools.PersonEmail{{
			Email:          "same@example.org",
			IsDefault:      true,
			ContactLabelID: 2,
		}},
	}

	plan := churchtools.PrepareEmailUpdate("same@example.org", person)
	if plan.Needed {
		t.Fatalf("expected no update, got %+v", plan)
	}
}

func TestPrepareEmailUpdateReplacePrimaryKeepOld(t *testing.T) {
	person := churchtools.Person{
		Email: "alt@example.org",
		Emails: []churchtools.PersonEmail{{
			Email:          "alt@example.org",
			IsDefault:      true,
			ContactLabelID: 2,
		}},
	}

	plan := churchtools.PrepareEmailUpdate("neu@example.org", person)
	if !plan.Needed {
		t.Fatal("expected update")
	}
	if plan.Primary != "neu@example.org" {
		t.Fatalf("primary = %q", plan.Primary)
	}
	if len(plan.Emails) != 2 {
		t.Fatalf("emails len = %d, want 2", len(plan.Emails))
	}
	if !plan.Emails[0].IsDefault || plan.Emails[0].Email != "neu@example.org" {
		t.Fatalf("first email = %+v", plan.Emails[0])
	}
	if plan.Emails[1].IsDefault || plan.Emails[1].Email != "alt@example.org" {
		t.Fatalf("second email = %+v", plan.Emails[1])
	}
}

func TestPrepareEmailUpdatePromoteExistingAdditional(t *testing.T) {
	person := churchtools.Person{
		Email: "alt@example.org",
		Emails: []churchtools.PersonEmail{
			{Email: "alt@example.org", IsDefault: true, ContactLabelID: 2},
			{Email: "neu@example.org", IsDefault: false, ContactLabelID: 3},
		},
	}

	plan := churchtools.PrepareEmailUpdate("neu@example.org", person)
	if !plan.Needed {
		t.Fatal("expected update")
	}
	if len(plan.Emails) != 2 {
		t.Fatalf("emails len = %d", len(plan.Emails))
	}
	for _, entry := range plan.Emails {
		if entry.Email == "neu@example.org" && !entry.IsDefault {
			t.Fatalf("neu should be default: %+v", plan.Emails)
		}
		if entry.Email == "alt@example.org" && entry.IsDefault {
			t.Fatalf("alt should not be default: %+v", plan.Emails)
		}
	}
}

func TestPrepareEmailUpdateSetMissingPrimary(t *testing.T) {
	person := churchtools.Person{Email: ""}
	plan := churchtools.PrepareEmailUpdate("neu@example.org", person)
	if !plan.Needed || plan.Primary != "neu@example.org" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}
