package docker

import (
	"net/http"
)

type Client struct {
	client *http.Client
	token  string
}

func New(client *http.Client) *Client {
	return &Client{
		client: client,
	}
}