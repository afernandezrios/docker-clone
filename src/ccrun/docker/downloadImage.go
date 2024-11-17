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

		authToken, err := client.Login(authInfo)
		if err != nil {
			log.Fatal(err)
		}

		manifest, err := client.PullManifestWithAuth(authToken.Token)
		if err != nil {
			log.Fatal(err)
		}

		client.DownloadLayers(*manifest, authToken.Token)
	}
}
