package opelogimpl

import (
	"io/fs"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func TestM2fBaseOpeLogs(t *testing.T) {
	lgr := common.DbgLogger()
	lgr.Debug("TestM2fOpeLogs: start")

	std := t.TempDir()
	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := MakeM2fManager(path.Join(ltd, "m2f.opl"))
	require.NoError(t, err)
	_, _, err = olm.Create("ds", "di", std, ttd)
	require.NoError(t, err)
	_, _, err = olm.Open("ds", "di", false)
	require.NoError(t, err)
	olm.PutLogicalEntry("", opelog.NewLogicalEntry())
	require.NoError(t, olm.Sync())
	olm.PutLogicalEntry("a", opelog.NewLogicalEntry())
	require.NoError(t, olm.Sync())
	olm2, err := MakeM2fManager(path.Join(ltd, "m2f.opl"))
	require.NoError(t, err)
	_, _, err = olm2.Open("ds", "di", false)
	require.Error(t, err)
	_, _, err = olm2.Open("ds", "di", true)
	require.NoError(t, err)
	require.NoError(t, olm.Close())
	require.NoError(t, olm2.Close())
	lgr.Debug("TestM2fOpeLogs: end")
}

func TestM2fWriteOpeLogs(t *testing.T) {
	lgr := common.DbgLogger()
	lgr.Debug("TestM2fOpeLogs: start")

	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 250, 25000, 2, 2025, false, 0))
	lgr.Debug("TestM2fOpeLogs: FileTreeGenerated")

	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := MakeM2fManager(path.Join(ltd, "m2f.opl"))
	require.NoError(t, err)
	sTs, _, err := olm.Create("ds", "di", std, ttd)
	require.NoError(t, err)
	_, _, err = olm.Open("ds", "di", false)
	require.NoError(t, err)
	err = filepath.Walk(std, func(path string, info fs.FileInfo, err error) error {
		le := opelog.NewLogicalEntry()
		se := &opelog.StoredEntry{IsDir: info.IsDir(), Mtime: info.ModTime().Unix()}
		le.CreateEvent(sTs, false, opelog.EVC_LOADED, se, nil)
		le.SetState(sTs, false, opelog.STC_DONE_PRESENT, se, nil, 0)
		return olm.PutLogicalEntry(path, le)
	})
	require.NoError(t, err)
	require.NoError(t, olm.Sync())
	olm2, err := MakeM2fManager(path.Join(ltd, "m2f.opl"))
	require.NoError(t, err)
	_, _, err = olm2.Open("ds", "di", false)
	require.Error(t, err)
	_, _, err = olm2.Open("ds", "di", true)
	require.NoError(t, err)
	require.NoError(t, olm.Close())
	require.NoError(t, olm2.Close())
	lgr.Debug("TestM2fOpeLogs: end")
}
