package churchtools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const defaultPersonPageSize = 100

// PersonListOptions filters persons loaded from ChurchTools.
type PersonListOptions struct {
	IDs     []int
	GroupID int
}

// ListPersons returns persons for export or batch processing.
func (c *Client) ListPersons(opts PersonListOptions) ([]Person, error) {
	if opts.GroupID > 0 {
		ids, err := c.groupMemberPersonIDs(opts.GroupID)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return nil, nil
		}
		opts.IDs = ids
	}

	if len(opts.IDs) > 0 {
		return c.listPersonsByIDs(opts.IDs)
	}

	return c.listAllPersons()
}

func (c *Client) listAllPersons() ([]Person, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(defaultPersonPageSize))

	items, err := c.fetchPersonPages("/persons", query)
	if err != nil {
		return nil, err
	}
	return decodePersonList(items)
}

func (c *Client) listPersonsByIDs(ids []int) ([]Person, error) {
	const chunkSize = 50
	persons := make([]Person, 0, len(ids))

	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]

		query := url.Values{}
		query.Set("limit", strconv.Itoa(len(chunk)))
		for _, id := range chunk {
			query.Add("ids[]", strconv.Itoa(id))
		}

		items, err := c.fetchPersonPages("/persons", query)
		if err != nil {
			return nil, err
		}

		chunkPersons, err := decodePersonList(items)
		if err != nil {
			return nil, err
		}
		persons = append(persons, chunkPersons...)
	}

	return sortPersonsByIDs(persons, ids), nil
}

func (c *Client) groupMemberPersonIDs(groupID int) ([]int, error) {
	path := "/groups/" + strconv.Itoa(groupID) + "/members"
	items, err := c.fetchPersonPages(path, nil)
	if err != nil {
		return nil, fmt.Errorf("gruppenmitglieder laden: %w", err)
	}

	seen := make(map[int]struct{}, len(items))
	ids := make([]int, 0, len(items))
	for _, item := range items {
		id, ok := personIDFromMember(item)
		if !ok || id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func (c *Client) fetchPersonPages(path string, query url.Values) ([]json.RawMessage, error) {
	if query == nil {
		query = url.Values{}
	}
	if query.Get("limit") == "" {
		query.Set("limit", strconv.Itoa(defaultPersonPageSize))
	}

	page := 1
	var all []json.RawMessage

	for {
		query.Set("page", strconv.Itoa(page))
		body, err := c.getAPI(path, query)
		if err != nil {
			return nil, err
		}

		var envelope struct {
			Data json.RawMessage `json:"data"`
			Meta struct {
				Pagination *struct {
					Current  int `json:"current"`
					LastPage int `json:"lastPage"`
				} `json:"pagination"`
			} `json:"meta"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			return nil, fmt.Errorf("antwort parsen: %w", err)
		}

		items, err := rawItems(envelope.Data)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)

		if envelope.Meta.Pagination == nil || page >= envelope.Meta.Pagination.LastPage {
			break
		}
		page++
	}

	return all, nil
}

func (c *Client) getAPI(path string, query url.Values) ([]byte, error) {
	reqURL := c.apiURL(path)
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.applyAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.relogin(); err != nil {
			return nil, err
		}
		return c.getAPI(path, query)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "daten konnten nicht geladen werden",
			Body:       string(body),
		}
	}
	return body, nil
}

func rawItems(data json.RawMessage) ([]json.RawMessage, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}

	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err == nil {
		return items, nil
	}

	return []json.RawMessage{data}, nil
}

func decodePersonList(items []json.RawMessage) ([]Person, error) {
	persons := make([]Person, 0, len(items))
	for _, item := range items {
		var person Person
		if err := json.Unmarshal(item, &person); err != nil {
			return nil, fmt.Errorf("person parsen: %w", err)
		}
		if person.ID <= 0 {
			continue
		}
		persons = append(persons, person)
	}
	return persons, nil
}

func personIDFromMember(raw json.RawMessage) (int, bool) {
	var member map[string]any
	if err := json.Unmarshal(raw, &member); err != nil {
		return 0, false
	}

	if id, ok := intFromAny(member["personId"]); ok {
		return id, true
	}
	if person, ok := member["person"].(map[string]any); ok {
		if id, ok := intFromAny(person["id"]); ok {
			return id, true
		}
	}
	return 0, false
}

func intFromAny(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return int(n), true
	default:
		return 0, false
	}
}

func sortPersonsByIDs(persons []Person, ids []int) []Person {
	byID := make(map[int]Person, len(persons))
	for _, person := range persons {
		byID[person.ID] = person
	}

	ordered := make([]Person, 0, len(ids))
	for _, id := range ids {
		if person, ok := byID[id]; ok {
			ordered = append(ordered, person)
		}
	}
	return ordered
}
