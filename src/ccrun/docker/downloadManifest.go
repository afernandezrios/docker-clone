package docker

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"runtime"
)

// Pull manifest for Alpine image without auth
func PullManifest() *http.Response {

	client := &http.Client{}
	alpineManifestPath := "https://registry.hub.docker.com/v2/alpine/git/manifests/latest"
	req, _ := http.NewRequest("GET", alpineManifestPath, nil)
	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("Cannot get authorization to pull alpine repository: %s \n", err)
	}

	return resp
}

// Pull manifest for Alpine image
func PullManifestWithAuth(token string) Manifest {

	client := &http.Client{}
	alpineManifestPath := "https://registry.hub.docker.com/v2/alpine/git/manifests/latest"

	req, _ := http.NewRequest("GET", alpineManifestPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("Cannot get authorization to pull alpine repository: %s \n", err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Cannot get image manifest: %s \n", resp.Status)
		return Manifest{}
	}

	manifestList := processManifestList(resp)
	validManifest := getManifestForCurrentOS(manifestList)

	return *validManifest
}

type ManifestListInfo struct {
	SchemaVersion int        `json:"schemaVersion"`
	MediaType     string     `json:"mediaType"`
	Manifests     []Manifest `json:"manifests"`
}

type Manifest struct {
	MediaType string   `json:"mediaType"`
	Digest    string   `json:"digest"`
	Size      int      `json:"size"`
	Platform  Platform `json:"platform"`
}

type Platform struct {
	Architecture string `json:"architecture"`
	Os           string `json:"os"`
}

func processManifestList(response *http.Response) (manifestList ManifestListInfo) {
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Read response failed, reason: %v \n", err)
	}

	manifestResponse := &ManifestListInfo{}
	if err := json.Unmarshal(body, &manifestResponse); err != nil {
		log.Fatalf("Parse response failed, reason: %v \n", err)
	}

	return *manifestResponse
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
