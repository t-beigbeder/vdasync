package cli

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseUrl(t *testing.T) {
	var (
		pluginName string
		host       string
		port       int
		rootPath   string
		err        error
	)
	type testDataType struct {
		url        string
		pluginName string
		host       string
		port       int
		rootPath   string
		hasErr     bool
	}
	for _, ref := range []testDataType{
		{"relativeRootPath", "", "", 0, "relativeRootPath", false},
		{"/rootPath1", "", "", 0, "/rootPath1", false},
		{"/root/Path2", "", "", 0, "/root/Path2", false},
		{"dss:/rootPath3", "", "", 0, "/rootPath3", false},
		{"dss:/root/Path4", "", "", 0, "/root/Path4", false},
		{"lf+dss:/rootPath5", "lf", "", 0, "/rootPath5", false},
		{"lf+dss:/root/Path6", "lf", "", 0, "/root/Path6", false},
		{"dss://server/rootpath7", "", "server", 0, "/rootpath7", false},
		{"dss://server/root/path8", "", "server", 0, "/root/path8", false},
		{"dss://server:443/rootpath9", "", "server", 443, "/rootpath9", false},
		{"dss://server:443/root/path10", "", "server", 443, "/root/path10", false},
		{"notdss:/rootPath3", "", "", 0, "", true},
		{"lf+notdss:/rootPath3", "", "", 0, "", true},
		{"lf+notdss://server:443/root/path10", "", "", 0, "", true},
	} {
		pluginName, host, port, rootPath, err = ParseUrl(ref.url)
		require.Equal(t, ref.hasErr, err != nil)
		require.Equal(t, ref.pluginName, pluginName)
		require.Equal(t, ref.host, host)
		require.Equal(t, ref.port, port)
		if ref.rootPath != "relativeRootPath" {
			require.Equal(t, ref.rootPath, rootPath)
		} else {
			require.Equal(t, ref.rootPath, path.Base(rootPath))
		}
	}

	// relativeRootPath
	// /rootPath
	// [pluginName+]dss:/rootPath
	// [pluginName+]dss://host[:port]/rootPath

}
