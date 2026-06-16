package churchtools

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// InvitePerson sends the ChurchTools system invitation e-mail via REST API.
func (c *Client) InvitePerson(personID int) error {
	path := "/persons/" + strconv.Itoa(personID) + "/invite"
	req, err := http.NewRequest(http.MethodPost, c.apiURL(path), nil)
	if err != nil {
		return err
	}
	c.applyAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("einladung senden: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.relogin(); err != nil {
			return err
		}
		return c.InvitePerson(personID)
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	msg := inviteErrorMessage(resp.StatusCode, body)
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    msg,
		Body:       string(body),
	}
}

func inviteErrorMessage(status int, body []byte) string {
	switch status {
	case http.StatusBadRequest:
		return "person hat keine e-mail-adresse in ChurchTools"
	case http.StatusForbidden:
		return "keine berechtigung zum einladen (global invite person oder gruppe +invite persons)"
	case http.StatusNotFound:
		return "person nicht gefunden"
	default:
		if len(body) > 0 {
			return string(body)
		}
		return "einladung fehlgeschlagen"
	}
}
