package churchtools

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/janmz/churchtools-invite/internal/iolimit"
)

const maxOAuthRedirects = 15

// MeAPIToken returns the API login token for the current session. The
// endpoint is not part of the published OpenAPI spec; callers fall back to
// LoginToken when it fails.
func (c *Client) MeAPIToken() (string, error) {
	data, err := c.getJSONData("/person/me/apitoken", "API-Token konnte nicht gelesen werden")
	if err != nil {
		return "", err
	}
	var token string
	if err := json.Unmarshal(data, &token); err != nil {
		return "", fmt.Errorf("api-token parsen: %w", err)
	}
	return strings.TrimSpace(token), nil
}

func (c *Client) authenticateSubInstanceViaOAuth(centralURL, subURL string) error {
	if err := c.postPasswordLogin(centralURL); err != nil {
		return fmt.Errorf("login zentralinstanz: %w", err)
	}

	startURL, authorizeURL, err := discoverOAuthStartLogin(c.http, subURL)
	if err != nil {
		return err
	}
	// The sub-instance must trust exactly the configured main instance;
	// otherwise the central session would be used for an unknown host.
	if !sameOrigin(authorizeURL, centralURL) {
		return fmt.Errorf("nebeninstanz leitet die Anmeldung an %s weiter, konfiguriert ist %s – main_url prüfen", originOf(authorizeURL), centralURL)
	}

	if err := c.followOAuthRedirects(startURL, []string{subURL, centralURL}); err != nil {
		return err
	}

	c.baseURL = strings.TrimSuffix(subURL, "/")
	return c.finishPasswordLogin()
}

// DiscoverMainInstanceURL asks the sub-instance which central instance it
// uses for OAuth logins (the Location of /oauthclients/{id}/startlogin). No
// credentials are sent. This is the host the sub-instance itself trusts,
// unlike a host derived from the instance name.
func DiscoverMainInstanceURL(subURL string) (string, error) {
	httpClient := &http.Client{Timeout: 15 * time.Second}
	_, authorizeURL, err := discoverOAuthStartLogin(httpClient, normalizeInstanceURL(subURL))
	if err != nil {
		return "", err
	}
	if !allowedScheme(authorizeURL) {
		return "", fmt.Errorf("hauptinstanz %s verwendet kein https", originOf(authorizeURL))
	}
	mainURL := originOf(authorizeURL)
	if sameOrigin(authorizeURL, subURL) {
		return "", errors.New("nebeninstanz meldet keine eigene hauptinstanz")
	}
	return mainURL, nil
}

// discoverOAuthStartLogin finds the OAuth start URL of the sub-instance and
// the central authorize URL it redirects to. Redirects are not followed.
func discoverOAuthStartLogin(httpClient *http.Client, subURL string) (string, *url.URL, error) {
	noRedirect := &http.Client{
		Timeout: httpClient.Timeout,
		Jar:     httpClient.Jar,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	subURL = strings.TrimSuffix(subURL, "/")
	for id := 1; id <= 10; id++ {
		startURL := fmt.Sprintf("%s/oauthclients/%d/startlogin", subURL, id)
		req, err := http.NewRequest(http.MethodGet, startURL, nil)
		if err != nil {
			continue
		}

		resp, err := noRedirect.Do(req)
		if err != nil {
			continue
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, iolimit.MaxAPIResponseBytes))
		resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect:
			location, err := resp.Location()
			if err == nil && strings.Contains(location.Path, "/oauth/authorize") {
				return startURL, location, nil
			}
		}
	}

	return "", nil, errors.New("oauth-startlogin der nebeninstanz nicht gefunden")
}

// followOAuthRedirects runs the OAuth redirect chain. Only the given
// instances are allowed as redirect targets, and only via https (http only
// for localhost), so session cookies never go to another host or travel in
// plain text.
func (c *Client) followOAuthRedirects(startURL string, allowedInstances []string) error {
	client := &http.Client{
		Timeout: c.http.Timeout,
		Jar:     c.http.Jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxOAuthRedirects {
				return errors.New("zu viele oauth-redirects")
			}
			if !allowedScheme(req.URL) {
				return fmt.Errorf("oauth-weiterleitung ohne https abgelehnt: %s", originOf(req.URL))
			}
			for _, instance := range allowedInstances {
				if sameOrigin(req.URL, instance) {
					return nil
				}
			}
			return fmt.Errorf("oauth-weiterleitung zu %s abgelehnt (nur Haupt- und Nebeninstanz erlaubt)", originOf(req.URL))
		},
	}

	req, err := http.NewRequest(http.MethodGet, startURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("oauth-flow: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, iolimit.MaxAPIResponseBytes))
	return nil
}

// originOf returns scheme://host[:port] of u.
func originOf(u *url.URL) string {
	return u.Scheme + "://" + u.Host
}

// sameOrigin reports whether u has the scheme and host of instanceURL.
func sameOrigin(u *url.URL, instanceURL string) bool {
	other, err := url.Parse(normalizeInstanceURL(instanceURL))
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Scheme, other.Scheme) && strings.EqualFold(u.Host, other.Host)
}

// allowedScheme accepts https, and http only for localhost (tests).
func allowedScheme(u *url.URL) bool {
	switch strings.ToLower(u.Scheme) {
	case "https":
		return true
	case "http":
		host := strings.ToLower(u.Hostname())
		return host == "localhost" || host == "127.0.0.1" || host == "::1"
	}
	return false
}
