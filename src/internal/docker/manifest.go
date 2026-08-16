package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
)

const (
	BaseURL                     = "https://registry.hub.docker.com/v2/" // TODO: make it configurable
	ImageTag                    = "latest"                              // TODO: parse from cli input parameter
	MediaTypeDockerManifestList = "application/vnd.docker.distribution.manifest.v2+json"
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
	manifestPath := fmt.Sprintf("%s/%s/manifests/%s", BaseURL, imageName, ImageTag)

	req, err := http.NewRequest("GET", manifestPath, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Add("Accept", MediaTypeDockerManifestList)

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
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return decodeManifest(resp.Body)
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized: %w", NewUnAuthorizedError(*resp))
	default:
		// Drain body to allow TCP connection reuse
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("status received %w", ErrNotImplemented)
	}
}

func decodeManifest(manifestBody io.Reader) (*Manifest, error) {

	manifestResponse := &ManifestListInfo{}
	if err := json.NewDecoder(manifestBody).Decode(&manifestResponse); err != nil {
		return nil, fmt.Errorf("failed to decode manifest response: %v", err)
	}

	return getManifestForCurrentOS(*manifestResponse)
}

func getManifestForCurrentOS(manifestList ManifestListInfo) (*Manifest, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH

	for _, manifest := range manifestList.Manifests {
		if manifest.Platform.Os == os && manifest.Platform.Architecture == arch {
			return &manifest, nil
		}
	}
	return nil, fmt.Errorf("no matching manifest found for %s/%s", os, arch)
}
