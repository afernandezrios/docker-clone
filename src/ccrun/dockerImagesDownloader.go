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

// Login into docker registry to pull Alpine image.
func Login() (token string) {

	path := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull"

	loginPath := fmt.Sprintf(path, "alpine")

	resp, err := http.Get(loginPath)

	if err != nil {
		fmt.Printf("Cannot get authorization to pull alpine repository: %s \n", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Read response failed, reason: %v \n", err)
	}

	loginRespone := &LoginInfo{}
	if err := json.Unmarshal(body, &loginRespone); err != nil {
		log.Fatalf("Parse response failed, reason: %v \n", err)
	}

	fmt.Printf("New token: %s", loginRespone.Token)

	return loginRespone.Token
}
