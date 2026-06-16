package churchtools

import (
	"fmt"
)

const (
	ModuleChurchCore = "churchcore"
	ModuleChurchDB   = "churchdb"

	PermInvitePersons = "invite persons"
	PermExportData    = "export data"
	PermWriteAccess   = "write access"
)

// HasModulePermission checks a global permission flag from /permissions/global.
func HasModulePermission(perms map[string]any, module, name string) bool {
	mod, ok := perms[module].(map[string]any)
	if !ok {
		return false
	}
	value, ok := mod[name]
	if !ok {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case []any:
		return len(v) > 0
	case float64:
		return v != 0
	case int:
		return v != 0
	default:
		return false
	}
}

// HasModulePermission loads global permissions and checks one flag.
func (c *Client) HasModulePermission(module, name string) (bool, error) {
	perms, err := c.GlobalPermissions()
	if err != nil {
		return false, err
	}
	return HasModulePermission(perms, module, name), nil
}

// PermissionRequirement describes a missing permission and its grant group.
type PermissionRequirement struct {
	Module      string
	Permission  string
	GroupName   string
	Description string
}

// MembershipRequestStatus describes the outcome of a group membership request.
type MembershipRequestStatus string

const (
	MembershipActive    MembershipRequestStatus = "active"
	MembershipRequested MembershipRequestStatus = "requested"
	MembershipDenied    MembershipRequestStatus = "denied"
)

// MembershipRequestResult is returned by RequestGroupMembership.
type MembershipRequestResult struct {
	Status  MembershipRequestStatus
	Message string
}

// EnsurePermissions requests group membership when required permissions are missing.
func (c *Client) EnsurePermissions(reqs []PermissionRequirement) ([]string, error) {
	perms, err := c.GlobalPermissions()
	if err != nil {
		return nil, fmt.Errorf("berechtigungen laden: %w", err)
	}

	personID := c.PersonID()
	if personID <= 0 {
		user, err := c.WhoAmI()
		if err != nil {
			return nil, err
		}
		personID = user.ID
	}

	var notes []string
	requested := false

	for _, req := range reqs {
		if HasModulePermission(perms, req.Module, req.Permission) {
			continue
		}

		group, err := c.FindGroupByName(req.GroupName)
		if err != nil {
			notes = append(notes, fmt.Sprintf(
				"%s fehlt; gruppe %q nicht gefunden (%v)",
				req.Description,
				req.GroupName,
				err,
			))
			continue
		}

		result, err := c.RequestGroupMembership(group.ID, personID)
		if err != nil {
			notes = append(notes, fmt.Sprintf(
				"%s fehlt; anfrage für gruppe %q fehlgeschlagen: %v",
				req.Description,
				req.GroupName,
				err,
			))
			continue
		}
		requested = true

		switch result.Status {
		case MembershipActive:
			notes = append(notes, fmt.Sprintf(
				"%s fehlte; mitgliedschaft in %q hergestellt",
				req.Description,
				req.GroupName,
			))
		case MembershipRequested:
			notes = append(notes, fmt.Sprintf(
				"%s fehlt; mitgliedschaft in %q beantragt (freigabe durch administrator nötig)",
				req.Description,
				req.GroupName,
			))
		default:
			msg := result.Message
			if msg == "" {
				msg = "anfrage abgelehnt"
			}
			notes = append(notes, fmt.Sprintf(
				"%s fehlt; mitgliedschaft in %q nicht möglich: %s",
				req.Description,
				req.GroupName,
				msg,
			))
		}
	}

	if !requested {
		return notes, nil
	}

	perms, err = c.GlobalPermissions()
	if err != nil {
		return notes, err
	}

	for _, req := range reqs {
		if HasModulePermission(perms, req.Module, req.Permission) {
			continue
		}
		notes = append(notes, fmt.Sprintf(
			"%s weiterhin nicht aktiv (evtl. wartet freigabe)",
			req.Description,
		))
	}

	return notes, nil
}
