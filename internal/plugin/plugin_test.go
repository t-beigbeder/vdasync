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
	exep := path.Clean(testDir()+"/../../bin/testmain")
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
	cmds, errs := RunConfData(path.Join(t.TempDir(), "testgrpc.yml"), setExecutable(conf))
	if len(errs) > 0 {
		t.Error(errs)
	}
	require.Zero(t, len(errs))
	require.Equal(t, 1, len(cmds))
	errs = WaitFor(cmds)
	require.Zero(t, len(errs))
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
	cmds, errs := RunConfData(path.Join(t.TempDir(), "testgrpc.yml"), setExecutable(conf))
	if len(errs) > 0 {
		t.Error(errs)
	}
	require.Zero(t, len(errs))
	require.Equal(t, 2, len(cmds))
	errs = WaitFor(cmds)
	require.Zero(t, len(errs))
}

func TestRunPluginsOneMisConf(t *testing.T) {
	const conf string = `
plugins:
- name: localFilesSample1
  type: localFiles
  executablePath: ${exe}
  addArgs:
  - "-is-plugin"
- name: localFilesSample2
  type: localFiles
  executablePath: ${exe}doesnotexist
  addArgs:
  - "-is-plugin"
`
	cmds, errs := RunConfData(path.Join(t.TempDir(), "testgrpc.yml"), setExecutable(conf))
	require.Equal(t, 1, len(errs))
	require.Equal(t, 1, len(cmds))
	errs = WaitFor(cmds)
	require.Zero(t, len(errs))
}

func TestRunPluginsOneFail(t *testing.T) {
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
  - "-is-fatal"
`
	cmds, errs := RunConfData(path.Join(t.TempDir(), "testgrpc.yml"), setExecutable(conf))
	if len(errs) > 0 {
		t.Error(errs)
	}
	require.Zero(t, len(errs))
	require.Equal(t, 2, len(cmds))
	errs = WaitFor(cmds)
	require.Equal(t, 1, len(errs))
}
