package opelogimpl

import (
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func TestMoplmBaseOpeLogs(t *testing.T) {
	lgr := common.DbgLogger()
	lgr.Debug("TestMoplmOpeLogs: start")

	std := t.TempDir()
	ttd := t.TempDir()

	olm, err := MakeMemOplm()
	require.NoError(t, err)
	_, _, err = olm.Create("ds", "di", std, ttd)
	require.NoError(t, err)
	_, _, err = olm.Open("ds", "di", false)
	require.NoError(t, err)
	olm.PutLogicalEntry("", opelog.NewLogicalEntry())
	require.NoError(t, olm.Sync())
	olm.PutLogicalEntry("a", opelog.NewLogicalEntry())
	require.NoError(t, olm.Sync())
	_, _, err = olm.Open("ds", "di", false)
	require.Error(t, err)
	_, _, err = olm.Open("ds", "di", true)
	require.Error(t, err)
	require.NoError(t, olm.Close())
	_, _, err = olm.Open("ds", "di", false)
	require.NoError(t, olm.Close())
	lgr.Debug("TestMoplmOpeLogs: end")
}

func TestMoplmWriteOpeLogs(t *testing.T) {
	lgr := common.DbgLogger()
	lgr.Debug("TestMoplmOpeLogs: start")

	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 250, 25000, 2, 2025, false, 0))
	lgr.Debug("TestMoplmOpeLogs: FileTreeGenerated")

	ttd := t.TempDir()

	olm, err := MakeMemOplm()
	require.NoError(t, err)
	_, _, err = olm.Create("ds", "di", std, ttd)
	require.NoError(t, err)
	olm, err = MakeMemOplm()
	require.NoError(t, err)
	sTs, _, err := olm.Create("ds", "di", std, ttd)
	require.NoError(t, err)
	_, _, err = olm.Open("ds", "di", false)
	require.NoError(t, err)
	tSt := time.Now().Unix()
	err = filepath.Walk(std, func(path string, info fs.FileInfo, err error) error {
		le := opelog.NewLogicalEntry()
		se := &opelog.StoredEntry{IsDir: info.IsDir(), Mtime: info.ModTime().Unix()}
		le.CreateEvent(sTs, 0, false, opelog.EVC_LOADED, se, nil)
		le.SetState(tSt, sTs, false, opelog.STC_DONE_PRESENT, "", se, nil, 0)
		return olm.PutLogicalEntry(path, le)
	})
	require.NoError(t, err)
	require.NoError(t, olm.Sync())
	_, _, err = olm.Open("ds", "di", false)
	require.Error(t, err)
	require.NoError(t, olm.Close())
	_, _, err = olm.Open("ds", "di", true)
	require.NoError(t, err)
	require.NoError(t, olm.Close())
	lgr.Debug("TestMoplmOpeLogs: end")
}
