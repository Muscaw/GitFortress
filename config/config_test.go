package config

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func Test_loadConfig(t *testing.T) {
	configFolder, err := os.MkdirTemp("", "gitfortress_config_test")
	if err != nil {
		t.Fatalf("could not create temporary dir for config")
	}
	defer os.RemoveAll(configFolder)

	t.Run("no configuration file", func(t *testing.T) {
		viper.Reset()
		viper.AddConfigPath(configFolder)

		defer func() {
			if r := recover(); r != nil {
				if msg, ok := r.(error); ok {
					if !strings.Contains(msg.Error(), "could not find config file") {
						t.Fatalf("unexpected panic error returned: %v", msg.Error())
					}
				} else {
					t.FailNow()
				}
			} else {
				t.Fatalf("LoadConfig did not panic on missing configuration")
			}
		}()

		LoadConfig()

		// Should never come here as LoadConfig should panic
		t.FailNow()
	})

	t.Run("configuration with no inputs panics", func(t *testing.T) {
		viper.Reset()
		viper.AddConfigPath(configFolder)

		const missingInputs = "cloneFolderPath: /path/to/backup"

		err := os.WriteFile(path.Join(configFolder, "config.yml"), []byte(missingInputs), 0644)
		if err != nil {
			t.FailNow()
		}

		defer func() {
			if r := recover(); r != nil {
				if msg, ok := r.(error); ok {
					if !strings.Contains(msg.Error(), "could not validate config: expected to have at least one input") {
						t.Fatalf("unexpected panic error returned: %v", msg.Error())
					}
				} else {
					t.FailNow()
				}
			} else {
				t.Fatalf("LoadConfig did not panic on missing configuration")
			}

		}()

		LoadConfig()

		// Should never come here as LoadConfig should panic
		t.FailNow()

	})
	t.Run("configuration input has multiple times the same name", func(t *testing.T) {
		viper.Reset()
		viper.AddConfigPath(configFolder)

		const multipleInputWithSameNameConfig string = `---
inputs:
  github:
    - name: "Some input name"
      targetUrl: https://selfhosted.github.com
      apiToken: some-token
    - name: "Some input name"
      targetUrl: https://api.github.com
      apiToken: some-token
cloneFolderPath: /path/to/backup
ignoreRepositoriesRegex:
  - a-repo-name
influxDB:
  url: "http://influxurl"
  authToken: "influx_token"
  organizationName: "org_name"
  bucketName: "bucket_name"
prometheus:
  exposedPort: 1234
  autoConvertNames: false
`

		err := os.WriteFile(path.Join(configFolder, "config.yml"), []byte(multipleInputWithSameNameConfig), 0644)
		if err != nil {
			t.FailNow()
		}
		b, _ := os.ReadFile(path.Join(configFolder, "config.yml"))
		fmt.Print(string(b))

		defer func() {
			if r := recover(); r != nil {
				if msg, ok := r.(error); ok {
					if !strings.Contains(msg.Error(), "could not validate config: inputs must have unique names.") {
						t.Fatalf("unexpected panic error returned: %v", msg.Error())
					}
				} else {
					t.FailNow()
				}
			} else {
				t.Fatalf("LoadConfig did not panic on missing configuration")
			}
		}()

		LoadConfig()
		// Should never come here as LoadConfig should panic
		t.FailNow()
	})

	t.Run("configuration is parsed successfully", func(t *testing.T) {
		viper.Reset()
		viper.AddConfigPath(configFolder)

		const goodConfigFile string = `---
inputs:
  github:
    - name: "Some input name"
      targetUrl: https://api.github.com
      apiToken: some-token
      ignoreRepositoriesRegex:
        - a-repo-name
cloneFolderPath: /path/to/backup
influxDB:
  url: "http://influxurl"
  authToken: "influx_token"
  organizationName: "org_name"
  bucketName: "bucket_name"
prometheus:
  exposedPort: 1234
  autoConvertNames: false
`

		err := os.WriteFile(path.Join(configFolder, "config.yml"), []byte(goodConfigFile), 0644)
		if err != nil {
			t.FailNow()
		}
		b, _ := os.ReadFile(path.Join(configFolder, "config.yml"))
		fmt.Print(string(b))

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("LoadConfig should not panic. got %v", r)
			}
		}()

		config := LoadConfig()
		expectedConfig := Config{
			Inputs: Input{
				Github: []*GithubInput{{Name: "Some input name", TargetURL: "https://api.github.com", APIToken: "some-token", IgnoreRepositoriesRegex: []string{"a-repo-name"}}},
			},
			CloneFolderPath: "/path/to/backup",
			InfluxDB:        &InfluxDBConfig{Url: "http://influxurl", AuthToken: "influx_token", OrganizationName: "org_name", BucketName: "bucket_name"},
			Prometheus:      &PrometheusConfig{ExposedPort: 1234, AutoConvertNames: false},
			SyncDelay:       "5m",
		}

		if !reflect.DeepEqual(expectedConfig, config) {
			t.Fatalf("expected config does not match loaded config: %v %v", expectedConfig, config)
		}
	})
}
