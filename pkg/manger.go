package pkg

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	Candidate string   `json:"candidate"`
	Versions  []string `json:"versions"`
	Current   string   `json:"current"`
}

func getConfigPath() string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(userHome, DefaultLocation, DefaultConfigFile)
}

func LoadData() []Config {
	var configs []Config
	pathToCfg := getConfigPath()

	// check if file is created or not
	if _, err := os.Stat(pathToCfg); errors.Is(err, os.ErrNotExist) {
		return []Config{}
	} else {
		fileBytes, err := os.ReadFile(pathToCfg)

		if err != nil {
			panic(err)
		}

		if err := json.Unmarshal(fileBytes, &configs); err != nil {
			panic(err)
		}

		return configs
	}

	return nil
}

func saveData(configs []Config) {
	pathToCfg := getConfigPath()
	byteData, err := json.Marshal(configs)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(pathToCfg, byteData, 0755)
	if err != nil {
		panic(err)
	}
}

// UpdateDefaultVersionConfig will update current version for candidate -> Will be implemented later
func UpdateDefaultVersionConfig(candidate string, defaultVersion string) {
	configData := LoadData()
	for _, config := range configData {
		if config.Candidate == candidate {
			config.Current = defaultVersion
			break
		}
	}

	saveData(configData)
}

func AddNewVersion(candidate string, version string) {
	configData := LoadData()

	isExist := false
	for i, config := range configData {
		if config.Candidate == candidate {
			configData[i].Versions = append(configData[i].Versions, version)
			isExist = true
			break
		}
	}

	if !isExist {
		AddNewCandidate(candidate, version)
	} else {
		saveData(configData)
	}
}

func AddNewCandidate(candidate, version string) {
	configData := LoadData()
	newCandidate := Config{
		Candidate: candidate,
		Versions:  []string{version},
		Current:   version,
	}
	configData = append(configData, newCandidate)
	saveData(configData)
}
