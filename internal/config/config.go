package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

const configFilename = ".gatorconfig.json"

type Config struct {
	UrlDB           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (*Config, error) {
	config := Config{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("❗Failed to find home directory")
		return &Config{}, err
	}

	filePath := path.Join(homeDir, ".gatorconfig.json")
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❗Failed to find home directory")
		panic(err)
	}

	// parse file contents to json
	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("❗Failed to parse json string")
		return &Config{}, err
	}

	return &config, nil
}

func (c *Config) SetUser(name string) error {

}
