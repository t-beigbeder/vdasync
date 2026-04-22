package plugin

import (
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Dir(filename)
}

func setExecutable(conf string) string {
	exep := path.Clean(testDir() + "/../../bin/testmain")
	return strings.Replace(conf, "${exe}", exep, -1)
}

func TestRunPluginOk(t *testing.T) {
	const conf string = `
plugins:
- name: localFilesSample
  type: localFiles
  executablePath: ${exe}
  addArgs:
  - "-is-plugin"
`
	rps, err := RunConfData(path.Join(t.TempDir(), "TestRunPluginOk.yml"), setExecutable(conf))
	require.Nil(t, err)
	require.Equal(t, 1, len(rps))
	require.Zero(t, len(Errors(rps)))
	WaitFor(rps)
	require.Zero(t, len(Errors(rps)))
}

func TestRunPluginsOk(t *testing.T) {
	const conf string = `
plugins:
- name: localFilesSample1
  type: localFiles
  executablePath: ${exe}
  addArgs:
  - "-is-plugin"
- name: localFilesSample2
  type: localFiles
  executablePath: ${exe}
  addArgs:
  - "-is-plugin"
`
	rps, err := RunConfData(path.Join(t.TempDir(), "TestRunPluginsOk.yml"), setExecutable(conf))
	require.Nil(t, err)
	require.Equal(t, 2, len(rps))
	require.Zero(t, len(Errors(rps)))
	WaitFor(rps)
	require.Zero(t, len(Errors(rps)))
}

func TestRunPluginsOneMisConf(t *testing.T) {
	const conf string = `
plugins:
- name: localFilesSample1b
  type: localFiles
  executablePath: ${exe}
  addArgs:
  - "-is-plugin"
- name: localFilesSample2b
  type: localFiles
  executablePath: ${exe}doesnotexist
  addArgs:
  - "-is-plugin"
`
	rps, err := RunConfData(path.Join(t.TempDir(), "TestRunPluginsOneMisConf.yml"), setExecutable(conf))
	require.Nil(t, err)
	require.Equal(t, 2, len(rps))
	require.Equal(t, 1, len(Errors(rps)))
	WaitFor(rps)
	require.Equal(t, 1, len(Errors(rps)))
}

func TestRunPluginsOneFail(t *testing.T) {
	const conf string = `
plugins:
- name: localFilesSample1c
  type: localFiles
  executablePath: ${exe}
  addArgs:
  - "-is-plugin"
- name: localFilesSample2c
  type: localFiles
  executablePath: ${exe}
  addArgs:
  - "-is-plugin"
  - "-is-fatal"
`
	rps, err := RunConfData(path.Join(t.TempDir(), "TestRunPluginsOneFail.yml"), setExecutable(conf))
	require.Nil(t, err)
	require.Equal(t, 2, len(rps))
	require.Zero(t, len(Errors(rps)))
	WaitFor(rps)
	require.Equal(t, 1, len(Errors(rps)))
}
