package manifest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

type ManifestLayersData struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	Layers        []Layer `json:"layers"`
}

type Layer struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

func DownloadLayers(manifest Manifest, token string) (path string) {

	req, _ := http.NewRequest("GET", "https://registry.hub.docker.com/v2/alpine/git/manifests/"+manifest.Digest, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic(errors.New("Manifest layers not found: " + resp.Status))
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	layersData := &ManifestLayersData{}
	if err := json.Unmarshal(body, &layersData); err != nil {
		panic(err)
	}
	
	downloadPath := "../alpine-docker/"
	for _, layer := range layersData.Layers {
		downloadLayer(layer.Digest, token, downloadPath)
	}

	return downloadPath
}

// For simplicity, all documents will be downloaded in a hardcoded path and cache headers are ignored
// Doc: https://distribution.github.io/distribution/spec/api/#pulling-a-layer
func downloadLayer(digest string, token string, path string) {

	err := os.MkdirAll(path, 0700)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(path + digest)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	req, _ := http.NewRequest("GET", "https://registry.hub.docker.com/v2/alpine/git/blobs/"+digest, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic(errors.New("Layer blob not found: " + resp.Status))
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
}
