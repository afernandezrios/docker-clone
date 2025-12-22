package docker

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginData struct {
	Token string `json:"token"`
}

// Login into docker registry. Returns token.
// Expected input example:
// 	- realm:="https://auth.docker.io/token";
// 	- service:="registry.docker.io";
// 	- scope:="repository:alpine/git:pull";
// See https://distribution.github.io/distribution/spec/auth/token/#how-to-authenticate
func (c *Client) Login(authInfo AuthenticationInfo) (*LoginData, error) {

	req, err := http.NewRequest("GET", authInfo.Realm, nil)
	if err != nil {
		return nil, fmt.Errorf("Cannot create login request: %v", err)
	}

	reqUrl := req.URL
	queryParams := reqUrl.Query()
	queryParams.Set("service", authInfo.Service)
	queryParams.Set("scope", authInfo.Scope)
	reqUrl.RawQuery = queryParams.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to submit login request: %v", err)
	}

	var response *LoginData
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("Failed to decode login response: %v", err)
	}

	return response, nil
}
