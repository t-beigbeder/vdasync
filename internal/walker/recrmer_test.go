package walker

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/localfiles"
)

func TestActualRecRmer(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	_, _, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*256)
	require.Nil(t, err)
	ds := localfiles.MakeLocalFilesDssa()
	walker, err := RemoveAll(lgr, 4, ds, common.OsPath2DssPath(td1), false)
	require.Nil(t, err)
	de, err := ds.Stat(common.OsPath2DssPath(td1))
	require.NotNil(t, err)
	require.NotNil(t, de)
	require.True(t, de.ErrNotExist)

	rr := RmResult(walker)
	require.NotNil(t, rr)
}

func TestDryrunRecRmer(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	_, _, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*256)
	require.Nil(t, err)
	ds := localfiles.MakeLocalFilesDssa()

	walker := NewRecursiveRemover(lgr, 4, ds, true)
	de, err := ds.Stat(common.OsPath2DssPath(td1))
	require.Nil(t, err)
	err = walker.Run(de)
	require.Nil(t, err)

	rr := RmResult(walker)
	require.NotNil(t, rr)
}
