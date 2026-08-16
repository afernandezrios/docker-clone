package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var challengeRegex = regexp.MustCompile(`(.+)="(.+)"`)

type LoginData struct {
	Token string `json:"token"`
}

type AuthenticationInfo struct {
	Realm   string
	Service string
	Scope   string
}

func (c *Client) Authorize(authInfo AuthenticationInfo) error {
	loginResp, err := c.Login(authInfo)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	c.token = loginResp.Token
	return nil
}

// Login into docker registry. Returns token.
// Expected input example:
//   - realm:="https://auth.docker.io/token";
//   - service:="registry.docker.io";
//   - scope:="repository:alpine/git:pull";
//
// See https://distribution.github.io/distribution/spec/auth/token/#how-to-authenticate
func (c *Client) Login(authInfo AuthenticationInfo) (*LoginData, error) {

	req, err := buildLoginRequest(authInfo)
	if err != nil {
		return nil, err
	}

	return c.doLoginRequest(req)
}

func buildLoginRequest(authInfo AuthenticationInfo) (*http.Request, error) {
	req, err := http.NewRequest("GET", authInfo.Realm, nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create login request: %v", err)
	}

	reqUrl := req.URL
	queryParams := reqUrl.Query()
	queryParams.Set("service", authInfo.Service)
	queryParams.Set("scope", authInfo.Scope)
	reqUrl.RawQuery = queryParams.Encode()
	return req, nil
}

func (c *Client) doLoginRequest(req *http.Request) (*LoginData, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit login request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return extractLoginData(resp)

	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized: %w", NewUnAuthorizedError(*resp))

	default:
		// Drain body to allow TCP connection reuse
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("status received %w", ErrNotImplemented)
	}
}

func extractLoginData(response *http.Response) (*LoginData, error) {
	var loginData LoginData
	if err := json.NewDecoder(response.Body).Decode(&loginData); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %v", err)
	}

	return &loginData, nil
}

// ParseWwwAuthentication parse the authentication data received in a WWW-Authenticate header.
// The expected format input is: realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
//
// See also: [WWW-Authenticate doc].
//
// [WWW-Authenticate doc]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
func ParseWwwAuthentication(wwwAuthHeader string) (authInfo AuthenticationInfo) {

	// Example header: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:library/alpine:pull"
	authHeaderParts := strings.SplitN(strings.TrimSpace(wwwAuthHeader), " ", 2)
	if len(authHeaderParts) < 2 {
		return AuthenticationInfo{}
	}

	challenges := challengesExtractor(strings.Split(authHeaderParts[1], ","))

	return AuthenticationInfo{
		Realm:   challenges["realm"],
		Service: challenges["service"],
		Scope:   challenges["scope"],
	}
}

// expected input: realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
func challengesExtractor(headerValue []string) map[string]string {
	challenges := make(map[string]string)

	for _, value := range headerValue {
		match := challengeRegex.FindStringSubmatch(value)
		if match != nil {
			challenges[match[1]] = match[2]
		}
	}

	return challenges
}
