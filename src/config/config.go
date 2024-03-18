package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os/user"
	"path/filepath"
)

type Config struct {
	GithubToken             string   `yaml:"github_token" env:"GITFORTRESS_GITHUB_TOKEN" env-required:"true"`
	CloneFolderPath         string   `yaml:"clone_folder_path" env:"GITFORTRESS_CLONE_FOLDER_PATH"`
	IgnoreRepositoriesRegex []string `yaml:"ignore_repositories_regex" env:"GITFORTRESS_IGNORE_REPOSITORIES_REGEX" env-default:""`
	SyncDelay               string   `yaml:"sync_delay" env:"GITFORTRESS_SYNC_DELAY" env-default:"5m"`
	InfluxDBConfig          *struct {
		InfluxDBUrl       string `yaml:"url" env:"GITFORTRESS_INFLUXDB_URL" env-required:"true"`
		InfluxDBAuthToken string `yaml:"token" env:"GITFORTRESS_INFLUXDB_TOKEN" env-required:"false"`
		OrganizationName  string `yaml:"org_name" env:"GITFORTRESS_INFLUXDB_ORG_NAME" env-required:"true"`
		BucketName        string `yaml:"bucket_name" env:"GITFORTRESS_INFLUXDB_BUCKET_NAME" env-required:"true"`
	} `yaml:"influx_db"`
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
