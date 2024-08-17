package config

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type GithubInput struct {
	Name                    string
	TargetURL               string
	APIToken                string
	IgnoreRepositoriesRegex []string
}

func (i *GithubInput) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("input name must be set")
	}
	if i.TargetURL == "" {
		return fmt.Errorf("input targetUrl must be set")
	}
	if i.APIToken == "" {
		return fmt.Errorf("input apiToken must be set")
	}
	return nil
}

type Input struct {
	Github []GithubInput
}

func (i *Input) Validate() error {
	if len(i.Github) == 0 {
		return fmt.Errorf("expected to have at least one input")
	}
	inputNames := map[string]bool{}
	for _, g := range i.Github {
		g.Validate()
		if _, exists := inputNames[g.Name]; exists {
			return fmt.Errorf("inputs must have unique names. name %v appears at least twice", g.Name)
		}
		inputNames[g.Name] = true
	}
	return nil
}

type InfluxDBConfig struct {
	Url              string
	AuthToken        string
	OrganizationName string
	BucketName       string
}

func (i *InfluxDBConfig) Validate() error {
	if i.Url == "" {
		return fmt.Errorf("influx url must be set")
	}
	if i.AuthToken == "" {
		return fmt.Errorf("influx authToken must be set")
	}
	if i.OrganizationName == "" {
		return fmt.Errorf("influx organizationName must be set")
	}
	if i.BucketName == "" {
		return fmt.Errorf("influx bucketName must be set")
	}
	return nil
}

type PrometheusConfig struct {
	ExposedPort      int
	AutoConvertNames bool
}

func (p *PrometheusConfig) Validate() error {
	if p.ExposedPort == 0 {
		return fmt.Errorf("prometheus.exposedPort can not be 0")
	}
	return nil
}

type Config struct {
	Inputs          Input
	CloneFolderPath string
	SyncDelay       string
	InfluxDB        *InfluxDBConfig
	Prometheus      *PrometheusConfig
}

func (c *Config) Process() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if c.CloneFolderPath == "~" {
		c.CloneFolderPath = dir
	} else if strings.HasPrefix(c.CloneFolderPath, "~/") {
		c.CloneFolderPath = filepath.Join(dir, c.CloneFolderPath[2:])
	}
}

func (c *Config) Validate() error {
	if err := c.Inputs.Validate(); err != nil {
		return err
	}
	if c.CloneFolderPath == "" {
		return fmt.Errorf("CloneFolderPath is empty")
	}
	if c.SyncDelay == "" {
		return fmt.Errorf("SyncDelay is empty")
	}
	if c.InfluxDB != nil {
		if err := c.InfluxDB.Validate(); err != nil {
			return err
		}
	}
	if c.Prometheus != nil {
		if err := c.Prometheus.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func setDefaultValues() {
	viper.SetDefault("SyncDelay", "5m")
}

func init() {
	usr, _ := user.Current()
	homeDir := usr.HomeDir
	viper.AddConfigPath(filepath.Join(homeDir, ".config/gitfortress/"))
	viper.AddConfigPath("/etc/gitfortress/")
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	setDefaultValues()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("could not find config file: %w", err))
		}
		panic(fmt.Errorf("could not load config file: %w", err))
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("could not unmarshal configuration: %w", err))
	}
	err = config.Validate()
	if err != nil {
		panic(fmt.Errorf("could not validate config: %w", err))
	}
	config.Process()
	return config
}
