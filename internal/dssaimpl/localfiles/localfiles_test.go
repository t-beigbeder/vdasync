package localfiles

import (
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

func TestFileFunctions(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))
	lfd := MakeLocalFilesDssa()
	de, err := lfd.Stat(strings.Split(ft, "/"))
	require.Nil(t, err)
	require.Nil(t, common.Lutimes(ft, de.Mtime-600))
	de2, err := lfd.Stat(strings.Split(ft, "/"))
	require.Nil(t, err)
	require.Equal(t, de.Mtime-600, de2.Mtime)
}
