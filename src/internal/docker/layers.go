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
	"path/filepath"
	"strings"
)

type ManifestLayersData struct {
	SchemaVersion int        `json:"schemaVersion"`
	MediaType     string     `json:"mediaType"`
	Config        BlobInfo   `json:"config"`
	Layers        []BlobInfo `json:"layers"`
}

func (c *Client) DownloadLayers(manifest *Manifest, imageName string, downloadDir string) error {

	layerPath := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/manifests/%s", imageName, manifest.Digest)

	req, _ := http.NewRequest("GET", layerPath, nil)

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.makeGetLayersRequest(req, imageName, downloadDir)
}

func (c *Client) makeGetLayersRequest(req *http.Request, imageName string, downloadDir string) error {
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("cannot download image layers: %v", err)
	}

	switch resp.StatusCode {
	case 200:
		return c.extractLayers(resp, imageName, downloadDir)
	default:
		return fmt.Errorf("manifest layers not found: %s", resp.Status)
	}
}

func (c *Client) extractLayers(response *http.Response, imageName string, downloadDir string) error {
	defer response.Body.Close()

	layersData := &ManifestLayersData{}
	if err := json.NewDecoder(response.Body).Decode(&layersData); err != nil {
		return fmt.Errorf("failed to decode layers response: %v", err)
	}

	err := os.MkdirAll(downloadDir, 0777)
	if err != nil {
		return fmt.Errorf("cannot create directories to store image content: %v", err)
	}

	if err := c.downloadLayers(layersData, imageName, downloadDir); err != nil {
		return err
	}

	return c.downloadConfig(layersData, imageName, downloadDir)
}

func (c *Client) downloadConfig(layersData *ManifestLayersData, imageName string, downloadDir string) error {
	log.Printf("Downloading config in %s: %s\n", downloadDir, layersData.Config.Digest)
	configPath := downloadDir + "/config" + layersData.Config.GetExtension()
	return c.DownloadBlob(layersData.Config.Digest, configPath, imageName)
}

func (c *Client) downloadLayers(layersData *ManifestLayersData, imageName string, downloadDir string) error {

	for _, layer := range layersData.Layers {
		log.Printf("Downloading layer: %s\n", layer.Digest)

		filePath := downloadDir + layer.Digest + layer.GetExtension()

		if err := c.DownloadBlob(layer.Digest, filePath, imageName); err != nil {
			return err
		}

		if err := unzipLayer(filePath, downloadDir); err != nil {
			return err
		}
	}

	return nil
}

func unzipLayer(zipPath string, destPath string) error {

	file, err := os.Open(zipPath)

	if err != nil {
		return fmt.Errorf("error opening zip layer: %s", err)
	}

	defer file.Close()

	gzipReader, err := gzip.NewReader(file)

	if err != nil {
		return fmt.Errorf("error creating gzip reader: %s", err)
	}

	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break // End of archive
		}

		if err != nil {
			return fmt.Errorf("error iterating files in tar archive: %s", err)
		}

		handleLayerFile(header, destPath, tarReader)
	}

	return nil
}

func handleLayerFile(header *tar.Header, destPath string, tarReader *tar.Reader) {
	completeFilePath := header.Name
	fileName := filepath.Base(completeFilePath)

	// Documentation about whiteout files: https://specs.opencontainers.org/image-spec/layer/#whiteouts
	if fileName == ".wh..wh..opq" {
		handleOpaqueWhiteoutFile(completeFilePath, destPath)
		return
	}
	// Handle regular whiteout
	if strings.HasPrefix(fileName, ".wh.") {
		handleRegularWhiteoutFile(fileName, completeFilePath, destPath)
		return
	}

	newFilePath := filepath.Join(destPath, completeFilePath)
	handleCreationFile(header, newFilePath, tarReader, destPath)
}

func handleRegularWhiteoutFile(fileName string, completeFilePath string, destPath string) {
	log.Printf("Handling regular whiteout: %s\n", fileName)
	originalFileName := strings.TrimPrefix(fileName, ".wh.")
	pathToDelete := filepath.Join(destPath, filepath.Dir(completeFilePath), originalFileName)
	os.RemoveAll(pathToDelete)
}

func handleOpaqueWhiteoutFile(completeFilePath string, destPath string) {
	log.Printf("Handling opaque whiteout: %s\n", completeFilePath)
	dir := filepath.Join(destPath, filepath.Dir(completeFilePath))
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		os.RemoveAll(filepath.Join(dir, e.Name()))
	}
}

func handleCreationFile(header *tar.Header, newFilePath string, tarReader *tar.Reader, destPath string) {
	switch header.Typeflag {

	// Create directory
	case tar.TypeDir:
		if err := os.MkdirAll(newFilePath, os.FileMode(header.Mode)); err != nil {
			log.Printf("Directory %s cannot be created: %s\n", newFilePath, err)
		}

	// Create the file with the specified permissions
	case tar.TypeReg:
		outFile, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)

		if err != nil {
			log.Println("Error creating file:", err)
		}

		defer outFile.Close()

		// Copy file content
		if _, err := io.Copy(outFile, tarReader); err != nil {
			log.Println("Error copying file:", err)
		}

	// Create symlink
	case tar.TypeSymlink:
		target := header.Linkname
		// Important to remove existing file/dir/symlink created by previous layers
		if _, err := os.Lstat(newFilePath); err == nil {
			if err := os.RemoveAll(newFilePath); err != nil {
				log.Printf("Error removing existing path %s: %v\n", newFilePath, err)
				return
			}
		}

		if err := os.Symlink(target, newFilePath); err != nil {
			log.Printf("Error creating symlink %s -> %s: %v\n", newFilePath, target, err)
		}

	// Create hard link
	case tar.TypeLink:
		newFilePath = filepath.Join(destPath, header.Linkname)
		if err := os.Link(newFilePath, newFilePath); err != nil {
			log.Println("Error creating hard link:", err)
		}

	default:
		log.Printf("Unknown file type %c\n", header.Typeflag)
	}
}
