package churchtools

import (
	"net/http"
	"strconv"
)

// InvitePerson sends the ChurchTools system invitation e-mail via REST API
// (POST /persons/{id}/invite, HTTP 204 on success).
func (c *Client) InvitePerson(personID int) error {
	path := "/persons/" + strconv.Itoa(personID) + "/invite"
	status, body, err := c.callAPI(http.MethodPost, path, nil, nil)
	if err != nil {
		return err
	}
	if status == http.StatusNoContent || status == http.StatusOK {
		return nil
	}
	return newAPIError(status, inviteErrorMessage(status, body), body)
}

func inviteErrorMessage(status int, body []byte) string {
	switch status {
	case http.StatusBadRequest:
		// ChurchTools also answers 400 for other reasons (e.g. already a
		// user); its own text is appended by APIError.Error.
		if apiBodyMessage(body, "") != "" {
			return "Einladung abgelehnt"
		}
		return "Person hat keine E-Mail-Adresse in ChurchTools"
	case http.StatusForbidden:
		return "keine Berechtigung zum Einladen (global invite person oder Gruppe +invite persons)"
	case http.StatusNotFound:
		return "Person nicht gefunden"
	default:
		return apiBodyMessage(body, "Einladung fehlgeschlagen")
	}
}
