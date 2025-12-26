package docker

import (
	"errors"
	"fmt"

	"github.com/afernandezrios/docker-clone/internal/appctx"
)

// DownloadImage downloads all the necessary files related to an
// image: manifest, layers and config files.
func (c *Client) DownloadImage(ctx *context.Context) error {
	
    manifest, err := c.getManifest(ctx)

    if err != nil {
        return fmt.Errorf("pulling manifest: %w", err)
    }

    c.DownloadLayers(*manifest, ctx)
    return nil
}

func (c *Client) getManifest(ctx *context.Context) (*Manifest, error) {
    manifest, err := c.PullManifest(ctx)

    if err == nil {
        return manifest, nil
    }

	var unauthErr UnauthorizedError
	if errors.As(err, &unauthErr) {
		authInfo := ParseWwwAuthentication(unauthErr.authInfo)
		
        if err := c.Authorize(authInfo); err != nil {
            return nil, fmt.Errorf("authorization failed: %w", err)
        }

		return c.PullManifest(ctx)
	}

    return nil, fmt.Errorf("unexpected error: %w", err)
}