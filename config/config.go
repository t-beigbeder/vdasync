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
	ToBeTested     string   `yaml:"toBeTested"`
}

type CliConfig struct {
	Version                    string        `yaml`
	Plugins                    []*PluginType `yaml`
	PluginReadyRetries         int           `yaml:"pluginReadyRetries"`
	PluginReadyTimeout         string        `yaml:"pluginReadyTimeout"`
	PluginAddress              string        `yaml:"pluginAddress"`
	PluginTransportCredentials string        `yaml:"pluginTransportCredentials"`
}

const CliConfigDefaultYaml string = `
version: "0.1"
plugins:
pluginReadyRetries: 3
pluginReadyTimeout: "20ms"
pluginAddress: "localhost"
pluginTransportCredentials: "insecure"
`

const PluginTypeDefaultYaml string = `
toBeTested: "shouldBeSet"
`

var defaultPluginTypeValues = &PluginType{}

func init() {
	yaml.Unmarshal([]byte(PluginTypeDefaultYaml), &defaultPluginTypeValues)
	exe, _ := os.Executable()
	defaultPluginTypeValues.ExecutablePath = exe
}

func umarshalPlugin(op *PluginType, b []byte) error {
	tp := &PluginType{}
	*tp = *defaultPluginTypeValues
	if err := yaml.Unmarshal(b, tp); err != nil {
		return err
	}
	*op = *tp
	return nil
}

func Load(configPath string) (*CliConfig, error) {
	conf := CliConfig{}
	if err := yaml.Unmarshal([]byte(CliConfigDefaultYaml), &conf); err != nil {
		return nil, err
	}
	if err := common.YamlLoad(configPath, &conf, yaml.CustomUnmarshaler(umarshalPlugin)); err != nil {
		return nil, err
	}
	return &conf, nil
}
