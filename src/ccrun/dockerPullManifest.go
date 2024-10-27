package ccrun

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

type FsLayer struct {
	BlobSum string `json:"blobSum"`
}

type ManifestInfo struct {
	Name      string    `json:"name"`
	Tag       string    `json:"tag"`
	FsLayers  []FsLayer `json:"fsLayers"`
	History   string    `json:"history"`
	Signature string    `json:"signature"`
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
		return PullManifestWithAuth(auth_token)
	} else if resp.StatusCode != 200 {
		fmt.Printf("Cannot get image manifest: %s \n", resp.Status)
		return ""
	} else {
		return processOkResponse(resp)
	}
}

// Pull manifest for Alpine image
func PullManifestWithAuth(token string) (signature string) {

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

	fmt.Printf("%s\n", body)

	// TODO: parse response
	return "manifestResponse.Signature"
}
