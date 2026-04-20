package config

import (
	"github.com/goccy/go-yaml"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

type PluginType struct {
	Name string `yaml`
	Type string `yaml`
	Port int    `yaml`
}

type CliConfig struct {
	Version   string       `yaml`
	Plugins   []PluginType `yaml`
	DummyTest string       `yaml:"dummyTest"`
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

`

func Load(configPath string) (*CliConfig, error) {
	conf := CliConfig{}
	if err := yaml.Unmarshal([]byte(DefaultCliYamlConfig), &conf); err != nil {
		return nil, err
	}
	if err := common.YamlLoad(configPath, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
