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
func (c *Client) DownloadBlob(digest string, blobPath string, imageName string) error {

	blobRequestPath := fmt.Sprintf("%s/%s/blobs/%s", BaseURL, imageName, digest)
	req, err := http.NewRequest("GET", blobRequestPath, nil)
	if err != nil {
		return fmt.Errorf("create blob request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.makeGetBlobRequest(req, blobPath)
}

func (c *Client) makeGetBlobRequest(req *http.Request, blobPath string) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot download blob: %s", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {

	case http.StatusOK:
		return c.saveBlob(resp.Body, blobPath)

	default:
		// Drain body to allow TCP connection reuse
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("blob not found: %s", resp.Status)
	}
}

func (*Client) saveBlob(src io.Reader, blobPath string) error {
	out, err := os.Create(blobPath)
	if err != nil {
		return fmt.Errorf("cannot create blob file: %s", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return fmt.Errorf("cannot copy blob in path '%s': %v", blobPath, err)
	}

	// TODO: verify digest checksum

	return nil
}

// Getting extension of the file according to: https://distribution.github.io/distribution/spec/manifest-v2-2/#media-types
// It also supports OCI standard -> https://github.com/opencontainers/image-spec/blob/main/manifest.md#image-manifest-property-descriptions
func (blob BlobInfo) GetExtension() string {
	switch blob.MediaType {

	// “Layer”, as a gzipped tar
	case "application/vnd.docker.image.rootfs.diff.tar.gzip",
		"application/vnd.oci.image.layer.v1.tar+gzip":
		return ".tar.gz"

	// Container config JSON
	case "application/vnd.docker.container.image.v1+json",
		"application/vnd.oci.image.config.v1+json":
		return ".json"

	default:
		log.Println("MediaType not implemented: " + blob.MediaType)
		return ""
	}
}
