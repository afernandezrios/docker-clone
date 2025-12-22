package docker

import (
	"errors"
    "fmt"
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

// DownloadImage downloads all the necessary files related to an
// image: manifest, layers and config files.
func (c *Client) DownloadImage(downloadPath, imageName string) error {
	
    manifest, err := c.pullManifest(imageName)

    if err != nil {
        return fmt.Errorf("pulling manifest: %w", err)
    }

    c.DownloadLayers(*manifest, downloadPath, imageName)
    return nil
}

func (c *Client) pullManifest(imageName string) (*Manifest, error) {
    manifest, err := c.PullManifest(imageName)

    if err == nil {
        return manifest, err
    }

	var unauthErr UnauthorizedError
	if errors.As(err, &unauthErr) {
		authInfo := ParseWwwAuthentication(unauthErr.authInfo)
		
        if err := c.Authorize(authInfo); err != nil {
            return nil, fmt.Errorf("authorization failed: %w", err)
        }

		return c.PullManifestWithAuth(imageName)
	}  else {
        return nil, fmt.Errorf("unexpected error: %w", err)
    }
}

func (c *Client) Authorize(authInfo AuthenticationInfo) error {
	loginResp, err := c.Login(authInfo)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	c.token = loginResp.Token
	return nil
}
