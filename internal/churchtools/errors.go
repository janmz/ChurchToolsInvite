package churchtools

import (
	"errors"
	"net/http"
)

// hasStatus reports whether err is an APIError with one of the codes.
func hasStatus(err error, codes ...int) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	for _, code := range codes {
		if apiErr.StatusCode == code {
			return true
		}
	}
	return false
}

// IsForbidden reports whether err is a ChurchTools 403 response.
func IsForbidden(err error) bool {
	return hasStatus(err, http.StatusForbidden)
}

// isUnauthorized reports a ChurchTools 401 response.
func isUnauthorized(err error) bool {
	return hasStatus(err, http.StatusUnauthorized)
}

// isInvalidRequest reports a ChurchTools 400 or 422 response.
func isInvalidRequest(err error) bool {
	return hasStatus(err, http.StatusBadRequest, http.StatusUnprocessableEntity)
}
