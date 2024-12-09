package docker

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
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

func (c *Client) DownloadLayers(manifest Manifest, downloadPath string) (path string) {

	req, _ := http.NewRequest("GET", "https://registry.hub.docker.com/v2/alpine/git/manifests/"+manifest.Digest, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Panicf("Cannot download image layers: %v", err)
	}

	if resp.StatusCode != 200 {
		log.Panicf("Manifest layers not found: %s", resp.Status)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("Cannot read layers response: %v", err)
	}

	layersData := &ManifestLayersData{}
	if err := json.Unmarshal(body, &layersData); err != nil {
		log.Panicf("Cannot parse layers response: %v", err)
	}

	err = os.MkdirAll(downloadPath, 0777)
	if err != nil {
		log.Panicf("Cannot create directories for the downloaded content: %v", err)
	}

	// Download layers
	for _, layer := range layersData.Layers {
		log.Printf("Downloading layer: %s\n", layer.Digest)
		filePath := downloadPath + layer.Digest + getExtension(layer)
		c.DownloadBlob(layer.Digest, filePath)
		unzipLayer(filePath, downloadPath)
	}

	// Download config
	log.Printf("Downloading config: %s\n", layersData.Config.Digest)
	configPath := downloadPath + "config" + getExtension(layersData.Config)
	c.DownloadBlob(layersData.Config.Digest, configPath)

	return downloadPath
}

// For simplicity, all documents will be downloaded in a hardcoded path and cache headers are ignored
// Doc: https://distribution.github.io/distribution/spec/api/#pulling-a-layer
func (c *Client) DownloadBlob(digest string, filePath string) string {

	out, err := os.Create(filePath)
	if err != nil {
		log.Panicf("Error creating layer file: %s", err)
	}
	defer out.Close()

	req, _ := http.NewRequest("GET", "https://registry.hub.docker.com/v2/alpine/git/blobs/"+digest, nil)
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
