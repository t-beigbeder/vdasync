package config

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

type PluginType struct {
	Name           string   `yaml`
	Type           string   `yaml`
	ExecutablePath string   `yaml:"executablePath"` // defaults to current executable
	AddArgs        []string `yaml:"addArgs"`
	Port           int      `yaml`
}

type CliConfig struct {
	Version   string        `yaml`
	Plugins   []*PluginType `yaml`
	DummyTest string        `yaml:"dummyTest"`
}

const DefaultCliYamlConfig string = `
version: "0.1"
plugins:
dummyTest: nothing
`

const CliYamlConfigSample string = `
plugins:
- name: localFilesSample
  type: localFiles
  port: 10314
  addArgs:
  - "-is-plugin"
`

func configurePlugins(config *CliConfig) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	for _, plugin := range config.Plugins {
		if plugin.ExecutablePath != "" {
			continue
		}
		plugin.ExecutablePath = exe
	}
	return nil
}

func Load(configPath string) (*CliConfig, error) {
	conf := CliConfig{}
	if err := yaml.Unmarshal([]byte(DefaultCliYamlConfig), &conf); err != nil {
		return nil, err
	}
	if err := common.YamlLoad(configPath, &conf); err != nil {
		return nil, err
	}
	if err := configurePlugins(&conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
