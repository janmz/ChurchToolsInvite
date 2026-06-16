package churchtools

import (
	"encoding/json"
	"fmt"
)

// PersonStatus is a ChurchTools person status (Mitglied, Gast, …).
type PersonStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ListPersonStatuses returns all person statuses.
func (c *Client) ListPersonStatuses() ([]PersonStatus, error) {
	items, err := c.fetchAPIList("/statuses", nil)
	if err != nil {
		return nil, fmt.Errorf("personenstatus laden: %w", err)
	}

	statuses := make([]PersonStatus, 0, len(items))
	for _, item := range items {
		var status PersonStatus
		if err := json.Unmarshal(item, &status); err != nil {
			return nil, fmt.Errorf("personenstatus parsen: %w", err)
		}
		if status.ID <= 0 {
			continue
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}
