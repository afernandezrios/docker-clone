package docker

import (
	"errors"
	"net/http"
)

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
