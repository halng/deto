package pkg

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halng/deto/tui"
	"io"
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

// Handler is an entry point for the package_manager.go file

func (man *Man) Handler() {
	// check if DefaultLocation exists or not
	homePath, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %s\n", err)
		os.Exit(1)
	}
	defaultLocation := filepath.Join(homePath, DefaultLocation)
	if _, err := os.Stat(defaultLocation); os.IsNotExist(err) {
		// get root directory
		err := os.MkdirAll(defaultLocation, 0777)
		if err != nil {
			fmt.Printf("Error creating directory: %s\n", err)
			os.Exit(1)
		}
	}

	switch man.ActionType {
	case "install":
		version := man.installNewVersion()
		AddNewVersion(man.Candidate, version)
	case "list":
		man.listOutAllVersion()

	default:
		fmt.Printf("Unsupported action type: %s\n", man.ActionType)
	}

}

func (man *Man) listOutAllVersion() {
	defaultCol := []string{
		"Version",
		"Current",
	}

	configData := LoadData()
	var rows [][]string

	for _, config := range configData {
		if config.Candidate == man.Candidate {
			for _, version := range config.Versions {
				isCurrent := ""
				if config.Current == version {
					isCurrent = "Current"
				}
				rows = append(rows, []string{version, isCurrent})
			}

		}
	}

	tui.Table(defaultCol, rows)

}

func (man *Man) installNewVersion() string {
	// handle business logic here.
	data := fetchRegistryData(*man)

	tui.Clear()
	listItem := make([]string, 0)
	for i, item := range data {
		listItem = append(listItem, fmt.Sprintf("%d| %s - %s - %s - Is LTS: %t", i+1, item.Name, item.Version, item.Provider, item.IsLTS))
	}
	title := "Select the version you want to install"
	selected := tui.InitList(listItem, title)

	idx, _ := strconv.Atoi(strings.Split(selected, "|")[0])

	if idx <= 0 {
		fmt.Println("You didn't select any item")
		os.Exit(1)
	}
	selectedItem := data[idx-1]
	// try to download and verify checksum
	isValid := DownloadAndVerify(selectedItem.Link, selectedItem.Checksum, "", selectedItem.Name)

	// extract the file
	if isValid {
		err := extractFile(selectedItem.Name, man.Candidate, selectedItem.Version)
		if err != nil {
			fmt.Printf("\nError: %s\n", err)
			os.Exit(1)
		}
		tui.Clear()
		fmt.Println("Installation completed")
	} else {
		fmt.Println("Checksum is not valid")
		os.Exit(1)
	}

	_ = os.Remove(selectedItem.Name)
	return selectedItem.Version
}

func fetchRegistryData(man Man) []RegistryVersion {
	msg := fmt.Sprintf("Starting checking data for OS: %s, Arch: %s", man.OperatingSystem, man.Architecture)
	modelSpinner := tui.InitialSpinnerModel()
	modelSpinner.Prompt = msg
	p := tea.NewProgram(modelSpinner)
	go func() {
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running spinner:", err)
			os.Exit(1)
		}
	}()

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

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		fmt.Println("Error reading response body:", err)
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
	p.Send(tea.Quit())
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
	tui.Clear()
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.ContentLength <= 0 {
		fmt.Println("can't parse content length, aborting download")
		os.Exit(1)
	}

	filePath := name
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var p *tea.Program
	pw := &tui.DownloadProgressWriter{
		Total:  int(resp.ContentLength),
		File:   file,
		Reader: resp.Body,
		OnProgress: func(ratio float64) {
			p.Send(tui.ProgressMsg(ratio))
		},
	}

	m := tui.DownloadProgressModel{
		Pw:       pw,
		Progress: progress.New(progress.WithDefaultGradient()),
	}

	p = tea.NewProgram(m)

	go pw.Start(p)

	if _, err := p.Run(); err != nil {
		fmt.Println("error running program:", err)
		os.Exit(1)
	}
	return filePath, nil
}

// verifyChecksum calculates the checksum of a file and compares it with the expected checksum
func verifyChecksum(filePath, expectedChecksum, algo string) (bool, error) {
	tui.Clear()
	msg := fmt.Sprintf("Verify checksum of %s", filePath)
	modelSpinner := tui.InitialSpinnerModel()
	modelSpinner.Prompt = msg
	p := tea.NewProgram(modelSpinner)
	go func() {
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running spinner:", err)
			os.Exit(1)
		}
	}()
	defer p.Send(tea.Quit())
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
	tui.Clear()
	userHome, _ := os.UserHomeDir()
	finalDest := filepath.Join(userHome, DefaultLocation, candidate, version)
	msg := fmt.Sprintf("Extracting from %s to %s ...", fileName, finalDest)
	modelSpinner := tui.InitialSpinnerModel()
	modelSpinner.Prompt = msg
	p := tea.NewProgram(modelSpinner)
	go func() {
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running spinner:", err)
			os.Exit(1)
		}
	}()

	defer p.Send(tea.Quit())

	if strings.Contains(fileName, ".tar.gz") {
		return decompressTarGz(fileName, finalDest)
	}
	return fmt.Errorf("unsupported file format: %s", fileName)
}

func decompressTarGz(src, finalDest string) error {
	// Open the tar.gz file

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
			if err := os.MkdirAll(targetPath, 0777); err != nil {
				return err
			}
		case tar.TypeReg:
			// Make the file and write its content
			if err := os.MkdirAll(filepath.Dir(targetPath), 0777); err != nil {
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
