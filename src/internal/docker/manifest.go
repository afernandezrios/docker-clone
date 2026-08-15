package docker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

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

func (c *Client) fetchManifest(imageName string) (*Manifest, error) {
	manifestPath := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/manifests/latest", imageName)

	req, _ := http.NewRequest("GET", manifestPath, nil)

	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.makeGetManifestRequest(req)
}

func (c *Client) makeGetManifestRequest(req *http.Request) (*Manifest, error) {
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("cannot pull image repository: %s", err)
	}

	switch status := resp.StatusCode; status {
	case 200:
		return extractManifest(resp)
	case 401:
		return nil, fmt.Errorf("unauthorized: %w", NewUnAuthorizedError(*resp))
	default:
		return nil, fmt.Errorf("status received %w", ErrNotImplemented)
	}
}

func extractManifest(response *http.Response) (manifest *Manifest, err error) {
	defer response.Body.Close()

	manifestResponse := &ManifestListInfo{}
	if err := json.NewDecoder(response.Body).Decode(&manifestResponse); err != nil {
		return nil, fmt.Errorf("failed to decode manifest response: %v", err)
	}

	return getManifestForCurrentOS(*manifestResponse), nil
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
