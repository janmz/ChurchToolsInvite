package churchtools

import (
	"errors"
	"fmt"
	"strings"
)

// PreJoinGroupResult describes one group in the pre-join sequence.
type PreJoinGroupResult struct {
	GroupName string
	Status    MembershipRequestStatus
	Message   string
	Skipped   bool
}

// EnsurePreJoinGroups joins the authenticated person to each group in order when
// not already a member. Groups that are not visible yet are retried in a
// further pass after a successful join (the new membership may grant the
// right to see them), so earlier list entries do not block later groups
// (e.g. Gruppen Administration before ChurchTools Admin). The person's
// groups are loaded once per pass.
func (c *Client) EnsurePreJoinGroups(groupNames []string) ([]PreJoinGroupResult, error) {
	personID := c.PersonID()
	if personID <= 0 {
		user, err := c.WhoAmI()
		if err != nil {
			return nil, err
		}
		personID = user.ID
	}

	names := make([]string, 0, len(groupNames))
	for _, name := range groupNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil, nil
	}

	completed := make(map[string]PreJoinGroupResult, len(names))
	pending := append([]string(nil), names...)

	for pass := 0; pass <= len(names) && len(pending) > 0; pass++ {
		var nextPending []string
		passProgress := false

		memberGroups, err := c.ListPersonGroups(personID)
		if err != nil {
			return orderedPreJoinResults(names, completed), fmt.Errorf("Gruppen der Person laden: %w", err)
		}

		for _, name := range pending {
			if groupsContainName(memberGroups, name) {
				completed[name] = PreJoinGroupResult{
					GroupName: name,
					Status:    MembershipActive,
					Message:   "Bereits Mitglied",
					Skipped:   true,
				}
				continue
			}

			group, err := c.FindGroupByName(name)
			if errors.Is(err, ErrGroupNotFound) {
				nextPending = append(nextPending, name)
				continue
			}
			if err != nil {
				return orderedPreJoinResults(names, completed), fmt.Errorf("Gruppe %q finden: %w", name, err)
			}

			membership, err := c.RequestGroupMembership(group.ID, personID)
			if err != nil {
				return orderedPreJoinResults(names, completed), fmt.Errorf("Gruppe %q beitreten: %w", name, err)
			}

			result := PreJoinGroupResult{
				GroupName: name,
				Status:    membership.Status,
				Message:   membership.Message,
				Skipped:   membership.AlreadyMember,
			}
			completed[name] = result
			if membership.Status == MembershipDenied {
				nextPending = append(nextPending, name)
			}
			// Only a newly active membership can grant rights that make other
			// groups visible or joinable; denials and pending requests do not,
			// so they must not trigger another pass.
			if membership.Status == MembershipActive && !membership.AlreadyMember {
				passProgress = true
			}
		}

		pending = nextPending
		if !passProgress {
			break // No progress made in this pass, so we're done
		}
	} // Next pass if not all groups were processed

	results := orderedPreJoinResults(names, completed)
	for i, result := range results {
		if result.Status != "" || result.Skipped {
			continue
		}
		results[i] = PreJoinGroupResult{
			GroupName: result.GroupName,
			Status:    MembershipDenied,
			Message:   "Gruppe nicht gefunden (evtl. erst nach vorheriger Gruppe sichtbar)",
		}
	}
	return results, nil
}

func orderedPreJoinResults(names []string, completed map[string]PreJoinGroupResult) []PreJoinGroupResult {
	results := make([]PreJoinGroupResult, 0, len(names))
	for _, name := range names {
		if result, ok := completed[name]; ok {
			results = append(results, result)
			continue
		}
		results = append(results, PreJoinGroupResult{GroupName: name})
	}
	return results
}
