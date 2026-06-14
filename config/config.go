package config

import (
	"github.com/goccy/go-yaml"
)

type PluginsOptionsType struct {
	NoTls          bool   `yaml:"noTls"`
	Insecure       bool   `yaml`
	CertPath       string `yaml:"certPath"`
	KeyPath        string `yaml:"keyPath"`
	CaCertPath     string `yaml:"caCertPath"`
	ClientCertPath string `yaml:"clientCertPath"`
	ClientKeyPath  string `yaml:"clientKeyPath"`
}

type PluginType struct {
	Name           string   `yaml`
	Type           string   `yaml`
	ExecutablePath string   `yaml:"executablePath"`
	AddArgs        []string `yaml:"addArgs"`
	Port           int      `yaml`
}

type DataStoreType struct {
	Host           string `yaml`
	Port           int    `yaml`
	Insecure       bool   `yaml`
	NoTls          bool   `yaml`
	CaCertPath     string `yaml:"caCertPath"`
	ClientCertPath string `yaml:"clientCertPath"`
	ClientKeyPath  string `yaml:"clientKeyPath"`
}

type SyncOptionsType struct {
	Dryrun  bool `yaml`
	Check   bool `yaml`
	NoPerm  bool `yaml:"noPerm"`
	NoMtime bool `yaml:"noMtime"`
	Rm      bool `yaml`
}

type CliConfig struct {
	Version            string              `yaml`
	PluginsOptions     *PluginsOptionsType `yaml:"pluginsOptions"`
	Plugins            []*PluginType       `yaml`
	PluginReadyRetries int                 `yaml:"pluginReadyRetries"`
	PluginReadyTimeout string              `yaml:"pluginReadyTimeout"`
	PluginAddress      string              `yaml:"pluginAddress"`
	DataStores         []*DataStoreType    `yaml:"dataStores"`
	SyncOptions        *SyncOptionsType    `yaml:"syncOptions"`
}

const CliConfigDefaultYaml string = `
version: "0.1"
pluginsOptions:
  noTls: false
  insecure: false
plugins:
pluginReadyRetries: 4
pluginReadyTimeout: "100ms"
pluginAddress: "localhost"
`

const PluginTypeDefaultYaml string = `
name: "default"
type: "localFiles"
`

var defaultPluginTypeValues = &PluginType{}

var DefaultPluginType = "localFiles"

func init() {
	yaml.Unmarshal([]byte(PluginTypeDefaultYaml), &defaultPluginTypeValues)
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

func Load(config string) (*CliConfig, error) {
	conf := CliConfig{}
	if err := yaml.Unmarshal([]byte(CliConfigDefaultYaml), &conf); err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalWithOptions([]byte(config), &conf, yaml.CustomUnmarshaler(umarshalPlugin)); err != nil {
		return nil, err
	}
	return &conf, nil
}

func RemoteDataStore(cfg *CliConfig, host string, port int) *DataStoreType {
	if cfg == nil {
		cfg = &CliConfig{}
	}
	for _, ds := range cfg.DataStores {
		if (ds.Host == "" || ds.Host == host) && ds.Port == port {
			return ds
		}
	}
	return nil
}
