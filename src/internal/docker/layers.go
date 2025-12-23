package docker

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type ManifestLayersData struct {
	SchemaVersion int        `json:"schemaVersion"`
	MediaType     string     `json:"mediaType"`
	Config        BlobInfo   `json:"config"`
	Layers        []BlobInfo `json:"layers"`
}

type BlobInfo struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

func (c *Client) DownloadLayers(manifest Manifest, downloadPath string, imageName string) (string, error) {

	layerPath := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/manifests/%s", imageName, manifest.Digest)
	
	req, _ := http.NewRequest("GET", layerPath, nil)

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.makeGetLayersRequest(req, downloadPath, imageName)
}

func (c *Client) makeGetLayersRequest(req *http.Request, downloadPath string, imageName string) (string, error) {
	resp, err := c.client.Do(req)

	if err != nil {
		return "", fmt.Errorf("cannot download image layers: %v", err)
	}

	switch resp.StatusCode {
	case 200:
		return c.extractLayers(resp, downloadPath, imageName)
	default:
		return "", fmt.Errorf("manifest layers not found: %s", resp.Status)
	}
}

func (c *Client) extractLayers(resp *http.Response, downloadPath string, imageName string) (string, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read layers response: %v", err)
	}

	layersData := &ManifestLayersData{}
	if err := json.Unmarshal(body, &layersData); err != nil {
		return "", fmt.Errorf("cannot parse layers response: %v", err)
	}

	err = os.MkdirAll(downloadPath, 0777)
	if err != nil {
		return "", fmt.Errorf("cannot create directories to store image content: %v", err)
	}

	c.downloadLayers(layersData, downloadPath, imageName)
	c.downloadConfig(layersData, downloadPath, imageName)

	return downloadPath, nil
}

func (c *Client) downloadConfig(layersData *ManifestLayersData, downloadPath string, imageName string) {
	log.Printf("Downloading config: %s\n", layersData.Config.Digest)
	configPath := downloadPath + "config" + getExtension(layersData.Config)
	c.DownloadBlob(layersData.Config.Digest, configPath, imageName)
}

func (c *Client) downloadLayers(layersData *ManifestLayersData, downloadPath string, imageName string) {
	for _, layer := range layersData.Layers {
		log.Printf("Downloading layer: %s\n", layer.Digest)
		filePath := downloadPath + layer.Digest + getExtension(layer)
		c.DownloadBlob(layer.Digest, filePath, imageName)
		unzipLayer(filePath, downloadPath)
	}
}

// For simplicity, all documents will be downloaded in a hardcoded path and cache headers are ignored
// Doc: https://distribution.github.io/distribution/spec/api/#pulling-a-layer
func (c *Client) DownloadBlob(digest string, filePath string, imageName string) string {

	out, err := os.Create(filePath)
	if err != nil {
		log.Panicf("Error creating layer file: %s", err)
	}
	defer out.Close()

	blobPath := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/blobs/%s", imageName, digest)
	req, _ := http.NewRequest("GET", blobPath, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Panicf("Error downloading layer: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Panicf("Layer blob not found: " + resp.Status)
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Panicf("Cannot copy blob in path '%s': %v", filePath, err)
	}
	return filePath
}

// Getting extension of the file according to: https://distribution.github.io/distribution/spec/manifest-v2-2/#media-types
// It also supports OCI standard -> https://github.com/opencontainers/image-spec/blob/main/manifest.md#image-manifest-property-descriptions
func getExtension(blobInfo BlobInfo) string {
	switch blobInfo.MediaType {
	case "application/vnd.docker.image.rootfs.diff.tar.gzip", "application/vnd.oci.image.layer.v1.tar+gzip":
		// “Layer”, as a gzipped tar
		return ".tar.gz"
	case "application/vnd.docker.container.image.v1+json", "application/vnd.oci.image.config.v1+json":
		// Container config JSON
		return ".json"
	default:
		log.Println("MediaType not implemented: " + blobInfo.MediaType)
		return ""
	}
}

func unzipLayer(zipPath string, destPath string) {
	// Open the gzipped tar file
	file, err := os.Open(zipPath)
	if err != nil {
		log.Panicf("Error opening zip layer: %s\n", err)
	}
	defer file.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		log.Panicf("Error creating gzip reader: %s\n", err)
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Panicf("Error iterating files in tar archive: %s\n", err)
		}

		// Handle the file based on its type
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(destPath+header.Name, os.FileMode(header.Mode)); err != nil {
				log.Printf("Directory %s cannot be created: %s\n", header.Name, err)
			}
		case tar.TypeReg:
			// Create the file with the specified permissions
			outFile, err := os.OpenFile(destPath+header.Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
			if err != nil {
				log.Println("Error creating file:", err)
			}
			defer outFile.Close()

			// Copy file content
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Println(err)
			}
		case tar.TypeSymlink:
			// Create symlink
			if err := os.Symlink(header.Linkname, destPath+header.Name); err != nil {
				log.Println(err)
			}
		case tar.TypeLink:
			// Create hard link
			if err := os.Link(destPath+header.Linkname, destPath+header.Name); err != nil {
				log.Println(err)
			}
		default:
			log.Printf("Unable to handle file type %c\n", header.Typeflag)
		}
	}
}
