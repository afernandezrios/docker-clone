package docker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

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

type LoginData struct {
	Token string `json:"token"`
}

// Login into docker registry. Returns token.
// Expected input example:
//   - realm:="https://auth.docker.io/token";
//   - service:="registry.docker.io";
//   - scope:="repository:alpine/git:pull";
//
// See https://distribution.github.io/distribution/spec/auth/token/#how-to-authenticate
func (c *Client) Login(authInfo AuthenticationInfo) (*LoginData, error) {

	req, err := http.NewRequest("GET", authInfo.Realm, nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create login request: %v", err)
	}

	reqUrl := req.URL
	queryParams := reqUrl.Query()
	queryParams.Set("service", authInfo.Service)
	queryParams.Set("scope", authInfo.Scope)
	reqUrl.RawQuery = queryParams.Encode()

	return c.doLoginRequest(req)
}

func (c *Client) doLoginRequest(req *http.Request) (*LoginData, error) {
	resp, err := c.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to submit login request: %v", err)
	}

	switch status := resp.StatusCode; status {
	case 200:
		return extractLoginData(resp)
	case 401:
		return nil, fmt.Errorf("unauthorized: %w", NewUnAuthorizedError(*resp))
	default:
		return nil, fmt.Errorf("status received %w", ErrNotImplemented)
	}
}

func extractLoginData(response *http.Response) (*LoginData, error) {
	defer response.Body.Close()

	var loginData *LoginData
	if err := json.NewDecoder(response.Body).Decode(&loginData); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %v", err)
	}

	return loginData, nil
}

// ParseWwwAuthentication parse the authentication data received in a WWW-Authenticate header.
// The expected format input is: realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
//
// See also: [WWW-Authenticate doc].
//
// [WWW-Authenticate doc]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
func ParseWwwAuthentication(wwwAuthHeader string) (authInfo AuthenticationInfo) {

	headerValue := getLastElement(strings.Split(wwwAuthHeader, " "))
	challenges := challengesExtractor(strings.Split(headerValue, ","))

	return AuthenticationInfo{
		Realm:   challenges["realm"],
		Service: challenges["service"],
		Scope:   challenges["scope"],
	}
}

func getLastElement(stringArray []string) string {
	return stringArray[len(stringArray)-1]
}

// expected input: realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
func challengesExtractor(headerValue []string) map[string]string {
	challenges := make(map[string]string)
	var challengeRegex = regexp.MustCompile(`(.+)="(.+)"`)

	for _, value := range headerValue {
		match := challengeRegex.FindStringSubmatch(value)
		log.Printf("Challenge added: %s\n", match[0])
		challenges[match[1]] = match[2]
	}

	return challenges
}
