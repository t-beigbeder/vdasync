package plugin

import (
	"context"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func testDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Dir(filename)
}

func setExecutable(ymlConf string) *config.CliConfig {
	exep := path.Clean(testDir() + "/../../bin/lamd64/localFiles")
	cf, _ := config.Load(strings.Replace(ymlConf, "${exe}", exep, -1))
	return cf
}

func TestRunPluginOk(t *testing.T) {
	const conf string = `
pluginsOptions:
  noTls: true
plugins:
- name: localFilesSample
  type: localFiles
  executablePath: ${exe}
  addArgs: [-notls, -log, stderr, -level, INFO]
`
	rps, err := RunCliConfig(common.GetLogger(), setExecutable(conf), nil)
	require.Nil(t, err)
	require.Equal(t, 1, len(rps))
	require.Zero(t, len(Errors(rps)))
	Shutdown(rps)
	require.Zero(t, len(Errors(rps)))
	WaitFor(rps)
	require.Zero(t, len(Errors(rps)))
}

func TestRunPluginsOk(t *testing.T) {
	const conf string = `
pluginsOptions:
  noTls: true
plugins:
- name: localFilesSample1
  type: localFiles
  executablePath: ${exe}
  addArgs: [-notls, -log, stderr, -level, INFO]
- name: localFilesSample2
  type: localFiles
  executablePath: ${exe}
  addArgs: [-notls, -log, stderr, -level, INFO]
`
	rps, err := RunCliConfig(common.GetLogger(), setExecutable(conf), nil)
	require.Nil(t, err)
	require.Equal(t, 2, len(rps))
	require.Zero(t, len(Errors(rps)))
	Shutdown(rps)
	require.Zero(t, len(Errors(rps)))
	WaitFor(rps)
	require.Zero(t, len(Errors(rps)))
}

func TestRunPluginsOneMisConf(t *testing.T) {
	const conf string = `
pluginsOptions:
  noTls: true
plugins:
- name: localFilesSample1b
  type: localFiles
  executablePath: ${exe}
  addArgs: [-notls, -log, stderr, -level, INFO]
- name: localFilesSample2b
  type: localFiles
  executablePath: ${exe}doesnotexist
  addArgs: [-notls, -log, stderr, -level, INFO]
`
	rps, err := RunCliConfig(common.GetLogger(), setExecutable(conf), nil)
	require.Nil(t, err)
	require.Equal(t, 2, len(rps))
	require.Equal(t, 1, len(Errors(rps)))
	Shutdown(rps)
	require.Equal(t, 1, len(Errors(rps)))
	WaitFor(rps)
	require.Equal(t, 1, len(Errors(rps)))
}

func TestRunPluginsOneFail(t *testing.T) {
	const conf string = `
pluginsOptions:
  noTls: true
plugins:
- name: localFilesSample1c
  type: localFiles
  executablePath: ${exe}
  addArgs: [-notls, -log, stderr, -level, INFO]
- name: localFilesSample2c
  type: localFiles
  executablePath: ${exe}
  addArgs: [-notls, -log, stderr, -level, INFO, -badoption]
`
	rps, err := RunCliConfig(common.GetLogger(), setExecutable(conf), nil)
	require.Nil(t, err)
	require.Equal(t, 2, len(rps))
	require.Equal(t, 1, len(Errors(rps)))
	Shutdown(rps)
	require.Equal(t, 1, len(Errors(rps)))
	WaitFor(rps)
	require.Equal(t, 1, len(Errors(rps)))
}

func TestRunPluginAndCallList(t *testing.T) {
	const conf string = `
pluginsOptions:
  noTls: true
plugins:
- name: localFilesSample
  type: localFiles
  executablePath: ${exe}
  addArgs: [-notls, -log, stderr, -level, INFO]
`
	rps, err := RunCliConfig(common.GetLogger(), setExecutable(conf), nil)
	require.Nil(t, err)
	require.Equal(t, 1, len(rps))
	require.Zero(t, len(Errors(rps)))
	rp := rps[0]
	wd, err := os.Getwd()
	require.Nil(t, err)
	des, err := rp.Client.List(context.Background(), &dssagrpc.Path{Path: wd})
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(des.Entries), 1)
	Shutdown(rps)
	require.Zero(t, len(Errors(rps)))
	WaitFor(rps)
	require.Zero(t, len(Errors(rps)))
}
