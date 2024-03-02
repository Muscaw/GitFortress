package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os/user"
	"path/filepath"
)

type Config struct {
	GithubToken        string   `yaml:"github_token" env:"GITFORTRESS_GITHUB_TOKEN" env-required:"true"`
	CloneFolderPath    string   `yaml:"clone_folder_path" env:"GITFORTRESS_CLONE_FOLDER_PATH"`
	IgnoreRepositories []string `yaml:"ignore_repositories_regex" env:"GITFORTRESS_IGNORE_REPOSITORIES" env-default:""`
	SyncDelay          string   `yaml:"sync_delay" env:"GITFORTRESS_SYNC_DELAY" env-default:"5m"`
}

func parseConfigFiles(configPaths ...string) (Config, error) {
	var config Config
	for _, configFile := range configPaths {
		err := cleanenv.ReadConfig(configFile, &config)
		if err == nil {
			return config, nil
		}
	}
	return config, fmt.Errorf("could not read config in the following locations: %v", configPaths)
}

func LoadConfig() Config {

	usr, _ := user.Current()
	homeDir := usr.HomeDir
	config, err := parseConfigFiles(filepath.Join(homeDir, ".config/gitfortress/config.yml"), "/etc/gitfortress/config.yml")
	if err != nil {
		panic(fmt.Sprintf("could not unmarshal configuration. reason: %+v", err))
	}
	return config
}
