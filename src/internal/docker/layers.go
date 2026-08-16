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

	layerPath := fmt.Sprintf("%s/%s/manifests/%s", BaseURL, imageName, manifest.Digest)

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
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return c.extractLayers(resp, imageName, downloadDir)
	default:
		return fmt.Errorf("manifest layers not found: %s", resp.Status)
	}
}

func (c *Client) extractLayers(response *http.Response, imageName string, downloadDir string) error {
	var layersData ManifestLayersData
	if err := json.NewDecoder(response.Body).Decode(&layersData); err != nil {
		return fmt.Errorf("failed to decode layers response: %v", err)
	}

	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("cannot create directories to store image content: %v", err)
	}

	if err := c.downloadLayers(&layersData, imageName, downloadDir); err != nil {
		return err
	}

	return c.downloadConfig(&layersData, imageName, downloadDir)
}

func (c *Client) downloadConfig(layersData *ManifestLayersData, imageName string, downloadDir string) error {
	log.Printf("Downloading config in %s: %s\n", downloadDir, layersData.Config.Digest)
	configPath := filepath.Join(downloadDir, "/config" + layersData.Config.GetExtension())
	return c.DownloadBlob(layersData.Config.Digest, configPath, imageName)
}

func (c *Client) downloadLayers(layersData *ManifestLayersData, imageName string, downloadDir string) error {

	for _, layer := range layersData.Layers {
		log.Printf("Downloading layer: %s\n", layer.Digest)
		filePath := filepath.Join(downloadDir, layer.Digest+layer.GetExtension())

		if err := c.DownloadBlob(layer.Digest, filePath, imageName); err != nil {
			return fmt.Errorf("download layer blob %s: %w", layer.Digest, err)
		}

		if err := unzipLayer(filePath, downloadDir); err != nil {
			return fmt.Errorf("unpack layer %s: %w", layer.Digest, err)
		}

		// Remove tarball after extracting to save disk space
		_ = os.Remove(filePath)
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

		if err := handleLayerFile(header, destPath, tarReader); err != nil {
			return fmt.Errorf("error handling layer file: %s", err)
		}
	}

	return nil
}

func handleLayerFile(header *tar.Header, destPath string, tarReader *tar.Reader) error {
	// Guard against Zip Slip / path traversal
	cleanedName := filepath.Clean(header.Name)
	targetPath := filepath.Join(destPath, cleanedName)

	rel, err := filepath.Rel(destPath, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path escapes destination: %s", header.Name)
	}

	baseName := filepath.Base(cleanedName)

	// Documentation about whiteout files: https://specs.opencontainers.org/image-spec/layer/#whiteouts
	if baseName == ".wh..wh..opq" {
		return handleOpaqueWhiteoutFile(targetPath)
	}
	if strings.HasPrefix(baseName, ".wh.") {
		return handleRegularWhiteoutFile(targetPath, baseName)
	}

	newFilePath := filepath.Join(destPath, cleanedName)
	return handleCreationFile(header, newFilePath, tarReader, destPath)
}

func handleRegularWhiteoutFile(targetPath, baseName string) error {
	origName := strings.TrimPrefix(baseName, ".wh.")
	pathToDelete := filepath.Join(filepath.Dir(targetPath), origName)
	return os.RemoveAll(pathToDelete)
}

func handleOpaqueWhiteoutFile(targetPath string) error {
	dir := filepath.Dir(targetPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func handleCreationFile(header *tar.Header, targetPath string, tarReader *tar.Reader, destPath string) error {
	switch header.Typeflag {

	// Create directory
	case tar.TypeDir:
		return os.MkdirAll(targetPath, os.FileMode(header.Mode))

	// Create the file with the specified permissions
	case tar.TypeReg:
		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			return err
		}
		defer outFile.Close()

		// Copy file content
		if _, err := io.Copy(outFile, tarReader); err != nil {
			return err
		}

		return nil

	// Create symlink
	case tar.TypeSymlink:
		// Important to remove existing file/dir/symlink created by previous layers
		_ = os.Remove(targetPath)
		return os.Symlink(header.Linkname, targetPath)

	// Create hard link
	case tar.TypeLink:
		oldFilePath := filepath.Join(destPath, header.Linkname)
		_ = os.Remove(targetPath)
		return os.Link(oldFilePath, targetPath)

	default:
		log.Printf("Unknown file type %c\n", header.Typeflag)
		return nil
	}
}
