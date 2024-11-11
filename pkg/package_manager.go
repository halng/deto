package pkg

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halng/deto/tui"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
Copyright © 2024 Hal Ng <haonguyentan2001@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// == In this package, we will manage the candidate and version that are installed in the system. == //

type Man struct {
	Candidate       string
	ActionType      string
	Architecture    string
	OperatingSystem string
}

type RegistryVersion struct {
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	Name         string `json:"name"`
	Checksum     string `json:"checksum"`
	Provider     string `json:"provider"`
	IsLTS        bool   `json:"is_lts"`
	Link         string `json:"link"`
}

type RegistryData struct {
	AIX       []RegistryVersion `json:"aix"`
	Darwin    []RegistryVersion `json:"darwin"`
	Linux     []RegistryVersion `json:"linux"`
	Dragonfly []RegistryVersion `json:"dragonfly"`
	Freebsd   []RegistryVersion `json:"freebsd"`
	Illumos   []RegistryVersion `json:"illumos"`
	Netbsd    []RegistryVersion `json:"netbsd"`
	Openbsd   []RegistryVersion `json:"openbsd"`
	Plan9     []RegistryVersion `json:"plan9"`
	Solaris   []RegistryVersion `json:"solaris"`
	Windows   []RegistryVersion `json:"windows"`
}

var DefaultLocation = "/.deto"

// Handler is an entry point for the package_manager.go file

func (man *Man) Handler() {
	fmt.Println("Starting package manager")
	// check if DefaultLocation exists or not
	homePath, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %s\n", err)
		os.Exit(1)
	}
	DefaultLocation = filepath.Join(homePath, DefaultLocation)
	fmt.Printf("Default location: %s\n", DefaultLocation)
	if _, err := os.Stat(DefaultLocation); os.IsNotExist(err) {
		// get root directory
		err := os.MkdirAll(DefaultLocation, 0777)
		if err != nil {
			fmt.Printf("Error creating directory: %s\n", err)
			os.Exit(1)
		}
	}

	// handle business logic here.
	data := fetchRegistryData(*man)
	for i, item := range data {
		fmt.Printf("\nItem: %d, \n\t Name: %s\n\t Version: %s\n\t Provider: %s\n\t LTS: %t", i, item.Name, item.Version, item.Provider, item.IsLTS)
	}
	idx, err := strconv.Atoi(tui.Input("Which item do you want to install?"))
	if err != nil {
		fmt.Printf("\nError: %s\n", err)
		os.Exit(1)
	}

	selectedItem := data[idx]
	// try to download and verify checksum
	isValid := DownloadAndVerify(selectedItem.Link, selectedItem.Checksum, "", selectedItem.Name)

	// extract the file
	if isValid {
		err = extractFile(selectedItem.Name, man.Candidate, selectedItem.Version)
		if err != nil {
			fmt.Printf("\nError: %s\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Checksum is not valid")
		os.Exit(1)
	}

	_ = os.Remove(selectedItem.Name)

}

func fetchRegistryData(man Man) []RegistryVersion {
	fmt.Printf("Staring checking data for OS: %s, Arch: %s", man.OperatingSystem, man.Architecture)
	url := fmt.Sprintf("https://raw.githubusercontent.com/halng/deto/refs/heads/main/registry/%s_versions.json", man.Candidate)

	resp, err := http.Get(url)

	if err != nil || resp != nil && resp.StatusCode == http.StatusNotFound {
		tea.Printf("Candidate: %s are not supported. Please try again later.", man.Candidate)
		os.Exit(1)

	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tea.Printf("Can not fetch data from registry. Error code %d", resp.StatusCode)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		os.Exit(1)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Error parsing response body:", err)
		os.Exit(1)
	}

	var result []RegistryVersion

	if osData, ok := data[man.OperatingSystem].([]interface{}); ok {
		for _, version := range osData {
			var registry RegistryVersion
			// Convert each `version` to JSON and then unmarshal it
			versionBytes, err := json.Marshal(version)
			if err != nil {
				fmt.Println("Error marshalling version:", err)
				os.Exit(1)
			}
			if err := json.Unmarshal(versionBytes, &registry); err != nil {
				fmt.Println("Error parsing registry version:", err)
				os.Exit(1)
			}
			result = append(result, registry)
		}
	}

	return result
}

func DownloadAndVerify(url string, checksum string, algo string, name string) bool {
	if algo == "" {
		algo = "sha256"
	}
	// Download the file
	filePath, err := downloadFile(url, name)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return false
	}

	// Verify checksum
	valid, err := verifyChecksum(filePath, checksum, algo)
	if err != nil {
		fmt.Println("Error verifying checksum:", err)
		return false
	}

	return valid
}

// downloadFile downloads a file from a URL and saves it locally
func downloadFile(url string, name string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filePath := name
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// verifyChecksum calculates the checksum of a file and compares it with the expected checksum
func verifyChecksum(filePath, expectedChecksum, algo string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	var hash []byte
	switch algo {
	case "sha256":
		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return false, err
		}
		hash = hasher.Sum(nil)
	case "sha512":
		hasher := sha512.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return false, err
		}
		hash = hasher.Sum(nil)
	default:
		return false, fmt.Errorf("unsupported hash algorithm: %s", algo)
	}

	// Convert hash to a hex string
	checksum := hex.EncodeToString(hash)
	return checksum == expectedChecksum, nil
}

func extractFile(fileName string, candidate string, version string) error {
	if strings.Contains(fileName, ".tar.gz") {
		return decompressTarGz(fileName, candidate, version)
	}
	return fmt.Errorf("unsupported file format: %s", fileName)
}

func decompressTarGz(src, candidate, version string) error {
	// Open the tar.gz file
	finalDest := filepath.Join(DefaultLocation, candidate, version)
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a gzip reader on the opened file
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// Create a tar reader on top of the gzip reader
	tarReader := tar.NewReader(gzipReader)

	// Iterate over the tar file entries
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		// Validate the file path
		if strings.Contains(header.Name, "..") {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		// Construct the full file path
		targetPath := filepath.Join(finalDest, header.Name)
		targetPath = filepath.Clean(targetPath)

		// Ensure the target path is within the final destination
		if !strings.HasPrefix(targetPath, filepath.Clean(finalDest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		// Check the type of entry
		switch header.Typeflag {
		case tar.TypeDir:
			// Make directory if not exists
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Make the file and write its content
			if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
				return err
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		default:
			log.Printf("Unable to handle file type %c in tar file", header.Typeflag)
		}
	}

	return nil
}
