package docker

import (
	"log"
)

// Pull manifest for Alpine image without auth
func DownloadImage() {

	resp := PullManifest()

	switch status := resp.StatusCode; status {
	case 401:
		// auth -> info in www-authenticate header: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
		authInfo := ParseWwwAuthentication(resp.Header.Get("www-authenticate"))
		auth_token := Login(authInfo)
		manifest := PullManifestWithAuth(auth_token)
		DownloadLayers(manifest, auth_token)
	default:
		log.Fatalf("Not implemented! Cannot get image manifest: %s \n", resp.Status)
	}
}
