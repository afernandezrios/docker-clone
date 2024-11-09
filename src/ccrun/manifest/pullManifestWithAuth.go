package manifest

import (
	"fmt"
	"net/http"
	"runtime"
)

// Pull manifest for Alpine image
func PullManifestWithAuth(token string) (Manifest) {

	client := &http.Client{}
	alpineManifestPath := "https://registry.hub.docker.com/v2/alpine/git/manifests/latest"

	req, _ := http.NewRequest("GET", alpineManifestPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Cannot get authorization to pull alpine repository: %s \n", err)
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Cannot get image manifest: %s \n", resp.Status)
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
