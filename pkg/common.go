package pkg

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadAndVerify(url string, checksum string, algo string) {
	if algo == "" {
		algo = "sha256"
	}
	// Download the file
	filePath, err := downloadFile(url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	defer os.Remove(filePath) // Clean up downloaded file after verification

	// Verify checksum
	valid, err := verifyChecksum(filePath, checksum, algo)
	if err != nil {
		fmt.Println("Error verifying checksum:", err)
		return
	}

	if valid {
		fmt.Println("Checksum verified successfully!")
	} else {
		fmt.Println("Checksum verification failed!")
	}
}

// downloadFile downloads a file from a URL and saves it locally
func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filePath := "downloaded_file.tar.gz"
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
