package docker

import (
	"errors"
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

// DownloadImage downloads all the neccessary files related to an 
// image: manifest, layers and config files.
func DownloadImage(downloadPath string) {

	client := NewClient(&http.Client{})
	_, err := client.PullManifest()

	var unauthErr UnauthorizedError
	if errors.As(err, &unauthErr) {
		authInfo := ParseWwwAuthentication(unauthErr.authInfo)
		client.Authorize(authInfo)
		manifest := client.PullManifestWithAuth()
		client.DownloadLayers(*manifest, downloadPath)
	}
}

func (c *Client) Authorize(authInfo AuthenticationInfo) {
	c.token = c.Login(authInfo).Token
}