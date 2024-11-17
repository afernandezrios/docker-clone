package docker

import (
	"log"
	"net/http"
)

// Pull manifest for Alpine image without auth
func DownloadImage() {

	client := NewClient(&http.Client{})

	resp := PullManifest()

	switch status := resp.StatusCode; status {
	case 401:
		// auth -> info in www-authenticate header: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
		authInfo := ParseWwwAuthentication(resp.Header.Get("www-authenticate"))
		authToken, err := client.Login(authInfo)
		if err != nil {
			log.Fatal(err)
		}
		// auth_token := Login(authInfo)
		manifest := PullManifestWithAuth(authToken.Token)
		DownloadLayers(manifest, authToken.Token)
	default:
		log.Fatalf("Not implemented! Cannot get image manifest: %s \n", resp.Status)
	}
}
