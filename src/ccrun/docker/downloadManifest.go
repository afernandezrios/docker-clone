package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
)

// Pull manifest for Alpine image without auth
func (c *Client) PullManifest() (*Manifest, error) {

	alpineManifestPath := "https://registry.hub.docker.com/v2/alpine/git/manifests/latest"
	req, _ := http.NewRequest("GET", alpineManifestPath, nil)
	resp, err := c.client.Do(req)

	if err != nil {
		log.Fatalf("Cannot get authorization to pull alpine repository: %s \n", err)
	}

	switch status := resp.StatusCode; status {
	case 401:
		// auth -> info in www-authenticate header: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
		// authInfo := ParseWwwAuthentication(resp.Header.Get("www-authenticate"))
		return nil, fmt.Errorf("unauthorized: %w", NewUnAuthorizedError(*resp))
	default:
		return nil, fmt.Errorf("status received %w", ErrNotImplemented)
	}
}

// Pull manifest for Alpine image
func (c *Client) PullManifestWithAuth(token string) (*Manifest, error) {

	alpineManifestPath := "https://registry.hub.docker.com/v2/alpine/git/manifests/latest"

	req, _ := http.NewRequest("GET", alpineManifestPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := c.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("cannot get authorization to pull alpine repository: %w", ErrInternalError)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Cannot get image manifest: %s \n", resp.Status)
		return nil, fmt.Errorf("status %w", ErrNotImplemented)
	}

	manifestList := processManifestList(resp)
	validManifest := getManifestForCurrentOS(manifestList)

	return validManifest, nil
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
