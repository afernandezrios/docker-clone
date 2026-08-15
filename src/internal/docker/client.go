package docker

import (
	"net/http"
)

type Client struct {
	httpClient *http.Client
	token      string
}

func New(client *http.Client) *Client {
	return &Client{
		httpClient: client,
	}
}
