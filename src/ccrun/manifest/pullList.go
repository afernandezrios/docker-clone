package manifest

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
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

func ProcessManifestList(response *http.Response) (manifestList ManifestListInfo) {
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
