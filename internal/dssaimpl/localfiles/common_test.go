package localfiles

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestDeList(t *testing.T) {
	dss := MakeLocalFilesDssa()
	td := t.TempDir()
	lgr := common.GetLogger()
	lgr = common.DbgLogger()
	_, _, err := common.AugmentTestFilesTree(td)
	require.NoError(t, err)
	des, err := dss.List(path.Join(td))
	require.NoError(t, err)
	for _, de := range des {
		lgr.Debug("TestDeList", "de", common.DataEntryList(de, true, ""))
	}
	des, err = dss.List(path.Join(td, "dLinks"))
	require.NoError(t, err)
	for _, de := range des {
		lgr.Debug("TestDeList", "de", common.DataEntryList(de, false, ""))
	}
	des, err = dss.List(path.Join(td, "dAddFiles"))
	require.NoError(t, err)
	for _, de := range des {
		lgr.Debug("TestDeList", "de", common.DataEntryList(de, false, ""))
	}
}

func TestMakeParents(t *testing.T) {
	dss := MakeLocalFilesDssa()
	td := t.TempDir()
	require.NoError(t, common.MakeParents(dss, path.Join(td, "d1a", "d2")))
	require.NoError(t, common.MakeParents(dss, path.Join(td, "d1b", "d3")))
}

func TestCopyEntry(t *testing.T) {
	td := t.TempDir()
	require.NoError(t, common.MakeTestFile(path.Join(td, "f1.txt"), 10*1024*1024))
	dss := MakeLocalFilesDssa()
	require.NoError(t, common.CopyEntry(dss, path.Join(td, "f1.txt"), path.Join(td, "f2.txt")))
	cs1, err := common.FileChecksum(path.Join(td, "f1.txt"), "")
	require.NoError(t, err)
	cs2, err := common.FileChecksum(path.Join(td, "f2.txt"), "")
	require.NoError(t, err)
	require.NotZero(t, len(cs2))
	require.Equal(t, cs1, cs2)
}

func TestChmod(t *testing.T) {
	dss := MakeLocalFilesDssa()
	td := t.TempDir()
	require.NoError(t, common.MakeParents(dss, path.Join(td, "d1a", "d2")))
	require.NoError(t, common.MakeParents(dss, path.Join(td, "d1b", "d3")))
	require.NoError(t, common.ChmodRO(dss, path.Join(td, "d1a")))
	require.NoError(t, common.ChmodRW(dss, path.Join(td, "d1a")))
	require.NoError(t, common.MakeTestFile(path.Join(td, "f1.txt"), 10))
	require.NoError(t, common.ChmodRO(dss, path.Join(td, "f1.txt")))
}

func TestLimBck(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	td := t.TempDir()
	dss := MakeLocalFilesDssa()
	path_ := path.Join(td, "TestLimBck.dat")
	require.NoError(t, common.MakeTestFile(path_, 1024))
	require.NoError(t, common.LimitedBackup(dss, path_, 0))
	time.Sleep(time.Second)
	require.NoError(t, common.MakeTestFile(path_, 1025))
	require.NoError(t, common.LimitedBackup(dss, path_, 0))
	time.Sleep(time.Second)
	require.NoError(t, common.MakeTestFile(path_, 1026))
	require.NoError(t, common.LimitedBackup(dss, path_, 0))
	pathFake := path.Join(td, "TestLimBck.dat."+common.BckExtFormat)
	require.NoError(t, common.MakeTestFile(pathFake, 1023))
	require.NoError(t, common.LimitedBackup(dss, path_, 2))
	time.Sleep(time.Second)
	require.NoError(t, common.MakeTestFile(path_, 1027))
	require.NoError(t, common.LimitedBackup(dss, path_, 2))
}
