package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

const CliYamlConfigSample string = `
plugins:
- name: localFilesSample
  type: localFiles
  port: 10314
  addArgs:
  - "-is-plugin"
`

func TestLoadConfig(t *testing.T) {
	td := t.TempDir()
	tf := filepath.Join(td, "TestLoadConfig.yml")
	common.WriteFile(tf, []byte(CliYamlConfigSample))
	config, err := Load(tf)
	if err != nil {
		t.Error(err)
	}
	require.Nil(t, err)
	require.Equal(t, "0.1", config.Version)
}
