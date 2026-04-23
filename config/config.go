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
	Version                    string        `yaml`
	Plugins                    []*PluginType `yaml`
	PluginReadyRetries         int           `yaml:"pluginReadyRetries"`
	PluginReadyTimeout         string        `yaml:"pluginReadyTimeout"`
	PluginAddress              string        `yaml:"pluginAddress"`
	PluginTransportCredentials string        `yaml:"pluginTransportCredentials"`
}

const DefaultCliYamlConfig string = `
version: "0.1"
plugins:
pluginReadyRetries: 3
pluginReadyTimeout: "20ms"
pluginAddress: "localhost"
pluginTransportCredentials: "insecure"
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
