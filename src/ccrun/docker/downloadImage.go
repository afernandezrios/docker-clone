package docker

import (
	"errors"
	"log"
	"net/http"
)

// Download image
func DownloadImage() {

	client := NewClient(&http.Client{})
	_, err := client.PullManifest()

	var unauthErr UnauthorizedError
	if errors.As(err, &unauthErr) {
		authInfo := ParseWwwAuthentication(unauthErr.authInfo)
		client.Authorize(authInfo)

		manifest, err := client.PullManifestWithAuth()
		if err != nil {
			log.Fatal(err)
		}

		client.DownloadLayers(*manifest)
	}
}
