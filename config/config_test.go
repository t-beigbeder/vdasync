package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

const CliConfigSample1Yaml string = `
plugins:
- name: localFilesSample
  type: localFiles
  port: 10314
  addArgs:
  - "-is-plugin"
`

const CliConfigSample2Yaml string = `
plugins:
- name: localFilesSample
  type: localFiles
  port: 10314
  addArgs:
  - "-is-plugin"
dataStores:
- name: localFileSystem
  type: localFiles
- name: pluginSample
  type: plugin
  pluginName: localFilesSample
- name: remoteSample
  type: remote
  host: localhost
  port: 10443
  tls: true
  caCertPath: x509/ca_cert.pem
  clientCertPath: x509/client_cert.pem
  clientKeyPath: x509/client_key.pem
`

func TestLoadConfig(t *testing.T) {
	td := t.TempDir()
	tf1 := filepath.Join(td, "TestLoadConfig1.yml")
	common.WriteFile(tf1, []byte(CliConfigSample1Yaml))
	config1, err := Load(tf1)
	if err != nil {
		t.Error(err)
	}
	require.Nil(t, err)
	require.Equal(t, "0.1", config1.Version)
	require.Equal(t, 1, len(config1.Plugins))
	require.Equal(t, "shouldBeSet", config1.Plugins[0].ToBeTested)

	tf2 := filepath.Join(td, "TestLoadConfig1.yml")
	common.WriteFile(tf2, []byte(CliConfigSample2Yaml))
	config2, err := Load(tf2)
	require.Nil(t, err)
	require.Equal(t, 3, len(config2.DataStores))
}
