package config

import (
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
)

var supportedInputTypes = []string{"github"}

func isInputTypeSupported(inputType string) bool {
	for _, supportedInputType := range supportedInputTypes {
		if supportedInputType == inputType {
			return true
		}
	}
	return false
}

type Input struct {
	Name                    string
	Type                    string
	TargetURL               string
	APIToken                string
	IgnoreRepositoriesRegex []string
}

func (i *Input) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("input name must be set")
	}
	if !isInputTypeSupported(i.Type) {
		return fmt.Errorf("input type is not supported: %v List of supported types: %v", i.Type, supportedInputTypes)
	}
	if i.TargetURL == "" {
		return fmt.Errorf("input targetUrl must be set")
	}
	if i.APIToken == "" {
		return fmt.Errorf("input apiToken must be set")
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
	Inputs          []Input
	CloneFolderPath string
	SyncDelay       string
	InfluxDB        *InfluxDBConfig
	Prometheus      *PrometheusConfig
}

func (c *Config) Validate() error {
	if len(c.Inputs) == 0 {
		return fmt.Errorf("expected to have at least one input")
	}
	inputNames := map[string]bool{}
	for _, i := range c.Inputs {
		if err := i.Validate(); err != nil {
			return err
		}
		if _, exists := inputNames[i.Name]; exists {
			return fmt.Errorf("inputs must have unique names. name %v appears at least twice", i.Name)
		}
		inputNames[i.Name] = true
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
	return config
}
