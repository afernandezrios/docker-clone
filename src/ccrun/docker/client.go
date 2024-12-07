package docker

import (
	"fmt"
	"net/http"
)

type Client struct {
	client *http.Client
	token string
}

func NewClient(client *http.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) Authorize(authInfo AuthenticationInfo) {
	var loginData *LoginData
	loginData, err := c.Login(authInfo)
	if err != nil {
		fmt.Printf("cannot authorize docker client")
	}

	c.token = loginData.Token
}
