package docker

import (
	"errors"
	"fmt"
)

// DownloadImage downloads all the necessary files related to an
// image: manifest, layers and config files.
func (c *Client) DownloadImage(imageName string, downloadDir string) error {

	manifest, err := c.pullManifest(imageName)
	if err != nil {
		return fmt.Errorf("pulling manifest: %w", err)
	}

	if err := c.DownloadLayers(manifest, imageName, downloadDir); err != nil {
		return fmt.Errorf("download layers for %q: %w", imageName, err)
	}

	return nil
}

func (c *Client) pullManifest(imageName string) (*Manifest, error) {
	manifest, err := c.fetchManifest(imageName)

	if err == nil {
		return manifest, nil
	}

	var unauthErr UnauthorizedError
	if errors.As(err, &unauthErr) {
		authInfo := ParseWwwAuthentication(unauthErr.authInfo)

		if err := c.Authorize(authInfo); err != nil {
			return nil, fmt.Errorf("authorization failed: %w", err)
		}

		manifest, err = c.fetchManifest(imageName)
		if err != nil {
			return nil, fmt.Errorf("retry manifest fetch: %w", err)
		}
		return manifest, nil
	}

	return nil, fmt.Errorf("unexpected error: %w", err)
}
