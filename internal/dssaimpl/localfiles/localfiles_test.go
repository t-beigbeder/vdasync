package localfiles

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
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
	de2.Mtime = de.Mtime
	de2.GroupRights = dssa.Rights{}
	de2.OtherRights = dssa.Rights{}
	err = lfd.SetStat(de2)
	require.Nil(t, err)
	de3, err := lfd.Stat(strings.Split(ft, "/"))
	require.Nil(t, err)
	require.Equal(t, de.Mtime, de3.Mtime)
	lt := path.Join(t.TempDir(), "TestFileFunctions.symlink")
	err = os.Symlink(ft, lt)
	require.Nil(t, err)
	ltde, err := lfd.Stat(strings.Split(lt, "/"))
	require.Nil(t, err)
	require.True(t, ltde.IsSymLink)
}
