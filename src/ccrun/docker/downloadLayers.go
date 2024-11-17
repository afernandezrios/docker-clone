package docker

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (c *Client) DownloadLayers(manifest Manifest, token string) (path string) {

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

	downloadPath := "../container-dir/"
	for _, layer := range layersData.Layers {
		fmt.Printf("Downloading layer: %s", layer.Digest)
		filePath := downloadLayer(layer, token, downloadPath)
		unzipLayer(filePath, downloadPath)
	}

	// Download config
	fmt.Printf("Downloading layer: %s", layersData.Config.Digest)
	downloadConfig(layersData.Config, token, downloadPath)

	return downloadPath
}

func downloadLayer(blobInfo BlobInfo, token string, path string) (filePath string) {

	err := os.MkdirAll(path, 0700)
	if err != nil {
		panic(err)
	}

	digest := blobInfo.Digest
	filePath = path + digest + getExtension(blobInfo)

	return downloadBlob(digest, token, filePath)
}

func downloadConfig(blobInfo BlobInfo, token string, path string) (filePath string) {

	err := os.MkdirAll(path, 0700)
	if err != nil {
		panic(err)
	}

	filePath = path + "config" + getExtension(blobInfo)
	return downloadBlob(blobInfo.Digest, token, filePath)
}

// For simplicity, all documents will be downloaded in a hardcoded path and cache headers are ignored
// Doc: https://distribution.github.io/distribution/spec/api/#pulling-a-layer
func downloadBlob(digest string, token string, filePath string) string {

	out, err := os.Create(filePath)
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
		fmt.Println("MediaType not implemented: " + blobInfo.MediaType)
		return ""
	}
}

func unzipLayer(zipPath string, destPath string) {
	// Open the gzipped tar file
	file, err := os.Open(zipPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
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
			panic(err)
		}

		// Handle the file based on its type
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(destPath+header.Name, os.FileMode(header.Mode)); err != nil {
				panic(err)
			}
		case tar.TypeReg:
			// Create file
			outFile, err := os.Create(destPath + header.Name)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer outFile.Close()

			// Copy file content
			if _, err := io.Copy(outFile, tarReader); err != nil {
				fmt.Println(err)
				return
			}
		case tar.TypeSymlink:
			// Create symlink
			if err := os.Symlink(destPath+header.Linkname, destPath+header.Name); err != nil {
				fmt.Println(err)
				return
			}
		case tar.TypeLink:
			// Create hard link
			if err := os.Link(destPath+header.Linkname, destPath+header.Name); err != nil {
				fmt.Println(err)
				return
			}
		default:
			fmt.Printf("Unable to handle file type %c\n", header.Typeflag)
		}
	}
}
