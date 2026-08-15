package docker

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type BlobInfo struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

// For simplicity, all documents will be downloaded in a hardcoded path and cache headers are ignored
// Doc: https://distribution.github.io/distribution/spec/api/#pulling-a-layer
func (c *Client) DownloadBlob(digest string, blobPath string, imageName string) (string, error) {

	blobRequestPath := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/blobs/%s", imageName, digest)
	req, _ := http.NewRequest("GET", blobRequestPath, nil)

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.makeGetBlobRequest(req, blobPath)
}

func (c *Client) makeGetBlobRequest(req *http.Request, blobPath string) (string, error) {
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return "", fmt.Errorf("cannot download blob: %s", err)
	}

	switch resp.StatusCode {
	case 200:
		return c.extractBlob(resp, blobPath)
	default:
		return "", fmt.Errorf("blob not found: %s", resp.Status)
	}
}

func (*Client) extractBlob(resp *http.Response, blobPath string) (string, error) {
	defer resp.Body.Close()

	out, err := os.Create(blobPath)

	if err != nil {
		return "", fmt.Errorf("cannot create blob file: %s", err)
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		return "", fmt.Errorf("cannot copy blob in path '%s': %v", blobPath, err)
	}

	return blobPath, nil
}

// Getting extension of the file according to: https://distribution.github.io/distribution/spec/manifest-v2-2/#media-types
// It also supports OCI standard -> https://github.com/opencontainers/image-spec/blob/main/manifest.md#image-manifest-property-descriptions
func (blob *BlobInfo) GetExtension() string {
	switch blob.MediaType {
	case "application/vnd.docker.image.rootfs.diff.tar.gzip", "application/vnd.oci.image.layer.v1.tar+gzip":
		// “Layer”, as a gzipped tar
		return ".tar.gz"
	case "application/vnd.docker.container.image.v1+json", "application/vnd.oci.image.config.v1+json":
		// Container config JSON
		return ".json"
	default:
		log.Println("MediaType not implemented: " + blob.MediaType)
		return ""
	}
}
