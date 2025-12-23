package docker

import (
	"errors"
    "fmt"
)

// DownloadImage downloads all the necessary files related to an
// image: manifest, layers and config files.
func (c *Client) DownloadImage(downloadPath, imageName string) error {
	
    manifest, err := c.getManifest(imageName)

    if err != nil {
        return fmt.Errorf("pulling manifest: %w", err)
    }

    c.DownloadLayers(*manifest, downloadPath, imageName)
    return nil
}

func (c *Client) getManifest(imageName string) (*Manifest, error) {
    manifest, err := c.PullManifest(imageName)

    if err == nil {
        return manifest, nil
    }

	var unauthErr UnauthorizedError
	if errors.As(err, &unauthErr) {
		authInfo := ParseWwwAuthentication(unauthErr.authInfo)
		
        if err := c.Authorize(authInfo); err != nil {
            return nil, fmt.Errorf("authorization failed: %w", err)
        }

		return c.PullManifest(imageName)
	}  else {
        return nil, fmt.Errorf("unexpected error: %w", err)
    }
}