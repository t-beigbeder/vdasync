package common

import (
	"github.com/stretchr/testify/require"
	"path"
	"testing"
)

type TestYamlFuncsEmbeddedStruct struct {
	D string `yaml`
}
type TestYamlFuncsStruct struct {
	A int      `yaml`
	B string   `yaml`
	C struct {
		D string `yaml`
	} `yaml`
}

func TestYamlFuncs(t *testing.T) {
	var err error
	ty := &TestYamlFuncsStruct{}
	err = YamlStore(path.Join(t.TempDir(), "TestYamlFuncs.yaml"), ty)
	require.Nil(t, err)
	var ty2 TestYamlFuncsStruct
	err = YamlLoad("../../testdata/internal/common/test.yaml", &ty2)
	require.Nil(t, err)
	ty3 := make(map[string]interface{})
	err = YamlLoad("../../testdata/internal/common/test.yaml", &ty3)
	require.Nil(t, err)
}
