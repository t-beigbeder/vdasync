package common

import (
	"github.com/stretchr/testify/require"
	"path"
	"testing"
)

type TestYamlFuncsStruct struct {
	A int      `yaml:"A"`
	B string   `yaml:"B"`
	C struct { // FIXME
		D string `yaml:"D"`
	}
}

func TestYamlFuncs(t *testing.T) {
	var err error
	ty := &TestYamlFuncsStruct{}
	err = YamlStore(path.Join(t.TempDir(), "TestYamlFuncs.yaml"), ty)
	require.Nil(t, err)
	var ty2 TestYamlFuncsStruct
	err = YamlLoad("../../testdata/internal/common/test.yaml", &ty2)
	require.Nil(t, err)
}
