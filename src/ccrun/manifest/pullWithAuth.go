package manifest

import (
	"errors"
	"log"
	"net/http"
	"runtime"

	"github.com/afernandezrios/docker-clone/ccrun/login"
)

// Pull manifest for Alpine image without auth
func PullManifest() (string, error) {

	client := &http.Client{}
	alpineManifestPath := "https://registry.hub.docker.com/v2/alpine/git/manifests/latest"
	req, _ := http.NewRequest("GET", alpineManifestPath, nil)
	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("Cannot get authorization to pull alpine repository: %s \n", err)
	}

	switch status := resp.StatusCode; status {
	case 401:
		// auth -> info in www-authenticate header: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
		authInfo := login.ParseWwwAuthentication(resp.Header.Get("www-authenticate"))
		auth_token := login.Login(authInfo)
		manifest := pullManifestWithAuth(auth_token)
		return DownloadLayers(manifest, auth_token), nil
	case 200:
		log.Fatalf("Cannot get image manifest without auth: %s \n", resp.Status)
		return "", errors.New("not implemented")
	default:
		log.Fatalf("Cannot get image manifest: %s \n", resp.Status)
		return "", errors.New("not implemented")
	}
}

// Pull manifest for Alpine image
func pullManifestWithAuth(token string) Manifest {

	client := &http.Client{}
	alpineManifestPath := "https://registry.hub.docker.com/v2/alpine/git/manifests/latest"

	req, _ := http.NewRequest("GET", alpineManifestPath, nil)
	req.Header.Set("Authorization", "Bearer "+ token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("Cannot get authorization to pull alpine repository: %s \n", err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Cannot get image manifest: %s \n", resp.Status)
		return Manifest{}
	}

	manifestList := ProcessManifestList(resp)
	validManifest := getManifestForCurrentOS(manifestList)

	return *validManifest
}

func getManifestForCurrentOS(manifestList ManifestListInfo) *Manifest {
	os := runtime.GOOS
	arch := runtime.GOARCH

	var validManifest *Manifest
	for _, manifest := range manifestList.Manifests {
		if manifest.Platform.Os == os && manifest.Platform.Architecture == arch {
			validManifest = &manifest
			break
		}
	}
	return validManifest
}
