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

	if err := c.downloadLayers(layersData, downloadPath, imageName); err != nil {
		return "", err
	}

	return c.downloadConfig(layersData, downloadPath, imageName)
}

func (c *Client) downloadConfig(layersData *ManifestLayersData, downloadPath string, imageName string) (string, error) {
	log.Printf("Downloading config: %s\n", layersData.Config.Digest)
	configPath := downloadPath + "config" + layersData.Config.GetExtension()
	return c.DownloadBlob(layersData.Config.Digest, configPath, imageName)
}

func (c *Client) downloadLayers(layersData *ManifestLayersData, downloadPath string, imageName string) error {

	for _, layer := range layersData.Layers {
		log.Printf("Downloading layer: %s\n", layer.Digest)

		filePath := downloadPath + layer.Digest + layer.GetExtension()

		if _, err := c.DownloadBlob(layer.Digest, filePath, imageName); err != nil {
			return err
		}

		if err := unzipLayer(filePath, downloadPath); err != nil {
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
	switch header.Typeflag {

	// Create directory
	case tar.TypeDir:
		if err := os.MkdirAll(destPath+header.Name, os.FileMode(header.Mode)); err != nil {
			log.Printf("Directory %s cannot be created: %s\n", header.Name, err)
		}

	// Create the file with the specified permissions
	case tar.TypeReg:
		outFile, err := os.OpenFile(destPath+header.Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)

		if err != nil {
			log.Println("Error creating file:", err)
		}

		defer outFile.Close()

		// Copy file content
		if _, err := io.Copy(outFile, tarReader); err != nil {
			log.Println(err)
		}

	// Create symlink
	case tar.TypeSymlink:
		if err := os.Symlink(header.Linkname, destPath+header.Name); err != nil {
			log.Println(err)
		}

	// Create hard link
	case tar.TypeLink:
		if err := os.Link(destPath+header.Linkname, destPath+header.Name); err != nil {
			log.Println(err)
		}

	default:
		log.Printf("Unknown file type %c\n", header.Typeflag)
	}
}
