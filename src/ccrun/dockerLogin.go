package ccrun

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type LoginInfo struct {
	Token string `json:"token"`
}

// Login into docker registry. Returns token.
// Expected input example:
// realm:="https://auth.docker.io/token";
// service:="registry.docker.io";
// scope:="repository:alpine/git:pull";
// See https://distribution.github.io/distribution/spec/auth/token/#how-to-authenticate
func Login(dockerAuthInfo AuthenticationInfo) (token string) {

	path := "%s?service=%s&scope=%s"

	loginPath := fmt.Sprintf(path, dockerAuthInfo.Realm, dockerAuthInfo.Service, dockerAuthInfo.Scope)

	resp, err := http.Get(loginPath)

	if err != nil {
		fmt.Printf("Cannot get authorization to pull alpine repository: %s \n", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Read response failed, reason: %v \n", err)
	}

	loginResponse := &LoginInfo{}
	if err := json.Unmarshal(body, &loginResponse); err != nil {
		log.Fatalf("Parse response failed, reason: %v \n", err)
	}

	return loginResponse.Token
}
