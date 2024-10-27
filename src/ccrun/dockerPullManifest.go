package ccrun

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
)

type ManifestInfo struct {
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

// Pull manifest for Alpine image wiuthout auth
func PullManifest() (signature string) {

	client := &http.Client{}
	alpineManifestPath := "https://registry.hub.docker.com/v2/alpine/git/manifests/latest"
	req, _ := http.NewRequest("GET", alpineManifestPath, nil)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Cannot get authorization to pull alpine repository: %s \n", err)
	}

	if resp.StatusCode == 401 {
		// auth -> info in www-authenticate header: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
		authInfo := ParseWwwAuthentication(resp.Header.Get("www-authenticate"))
		auth_token := Login(authInfo)
		return pullManifestWithAuth(auth_token)
	} else if resp.StatusCode != 200 {
		fmt.Printf("Cannot get image manifest: %s \n", resp.Status)
		return ""
	} else {
		return processOkResponse(resp)
	}
}

// Pull manifest for Alpine image
func pullManifestWithAuth(token string) (signature string) {

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
		return ""
	} else {
		return processOkResponse(resp)
	}
}

func processOkResponse(response *http.Response) (signature string) {
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Read response failed, reason: %v \n", err)
	}

	manifestResponse := &ManifestInfo{}
	if err := json.Unmarshal(body, &manifestResponse); err != nil {
		log.Fatalf("Parse response failed, reason: %v \n", err)
	}

	os := runtime.GOOS
	arch := runtime.GOARCH

	var validManifest *Manifest
	for _, manifest := range manifestResponse.Manifests {
		if manifest.Platform.Os == os && manifest.Platform.Architecture == arch {
			validManifest = &manifest
			break
		}
	}

	if validManifest == nil {
		fmt.Printf("Manifest not found for OS '%s' and architecture '%s'\n", os, arch)
		return ""
	}

	return validManifest.Digest
}
