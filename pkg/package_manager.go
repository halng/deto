package pkg

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halng/deto/tui"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
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
	// handle business logic here.
	data := CheckRegistryData(*man)
	for i, item := range data {
		fmt.Printf("\nItem: %d, \n\t Name: %s\n\t Version: %s\n\t Provider: %s\n\t LTS: %t", i, item.Name, item.Version, item.Provider, item.IsLTS)
	}
	idx, err := strconv.Atoi(tui.Input("Which item do you want to install?"))
	if err != nil {
		fmt.Printf("\nError: %s\n", err)
		os.Exit(1)
	}

	selectedItem := data[idx]
	DownloadAndVerify(selectedItem.Link, selectedItem.Checksum, "")
	// try to download and verify checksum

}

func CheckRegistryData(man Man) []RegistryVersion {
	fmt.Printf("Staring checking data for OS: %s, Arch: %s", man.OperatingSystem, man.Architecture)
	url := fmt.Sprintf("https://raw.githubusercontent.com/halng/deto/refs/heads/main/registry/%s_versions.json", man.Candidate)

	resp, err := http.Get(url)

	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			tea.Printf("Candidate: %s are not supported. Please try again later.", man.Candidate)
			os.Exit(1)
		}
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
