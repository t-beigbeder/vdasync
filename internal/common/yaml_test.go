package common

import (
	"path"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

type TestYamlFuncsEmbeddedStruct struct {
	F string `yaml`
	G string `yaml`
}
type TestYamlFuncsStruct struct {
	A int    `yaml`
	B string `yaml`
	C struct {
		D string `yaml`
	} `yaml`
	E TestYamlFuncsEmbeddedStruct `yaml`
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
	var ty4 TestYamlFuncsStruct
	err = YamlLoad("../../testdata/internal/common/test.yaml", &ty4,
		yaml.CustomUnmarshaler(func(otes *TestYamlFuncsEmbeddedStruct, b []byte) error {
			ttes := &TestYamlFuncsEmbeddedStruct{F: "default content for F", G: "default content for G"}
			if err := yaml.Unmarshal(b, ttes); err != nil {
				return err
			}
			*otes = *ttes
			return nil
		}))
	require.Nil(t, err)
}
