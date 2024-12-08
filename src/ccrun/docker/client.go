package docker

import (
	"net/http"
)

type Client struct {
	client *http.Client
	token  string
}

func NewClient(client *http.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) Authorize(authInfo AuthenticationInfo) {
	c.token = c.Login(authInfo).Token
}
