package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Acketuk/Gator/internal/database"
)

const configFilename = ".gatorconfig.json"

type Config struct {
	UrlDB           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}
type State struct {
	Db     *database.Queries
	Config *Config
}

//-------------------------------------------------------------------------------------

func Read() (*Config, error) {
	config := Config{}

	filePath, err := getConfigPath()
	if err != nil {
		fmt.Println(err)
		return &Config{}, err
	}
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❗Failed to open file - ", filePath)
		panic(err)
	}

	// parse json payload to Config struct
	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("❗Failed to parse json string")
		return &Config{}, err
	}

	return &config, nil
}

func (c *Config) SetUser(Name string) error {
	c.CurrentUserName = Name
	// writes json data to config file
	if err := write(c); err != nil {
		return err
	}

	return nil
}

//-------------------------- helpers ------------------------------------------------

func write(conf *Config) error {
	configJson, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		return err
	}
	configJson = append(configJson, '\n')

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	if err = os.WriteFile(configPath, configJson, 0644); err != nil {
		return err
	}

	return nil
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("⛔ unable to get $HOME directory")
		return "", err
	}
	configPath := filepath.Join(homeDir, configFilename)
	return configPath, nil
}
