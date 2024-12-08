package docker

import (
	"errors"
	"net/http"
)

func DownloadImage() {

	client := NewClient(&http.Client{})
	_, err := client.PullManifest()

	var unauthErr UnauthorizedError
	if errors.As(err, &unauthErr) {
		authInfo := ParseWwwAuthentication(unauthErr.authInfo)
		client.Authorize(authInfo)
		manifest := client.PullManifestWithAuth()
		client.DownloadLayers(*manifest)
	}
}
