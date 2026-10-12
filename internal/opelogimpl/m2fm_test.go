package opelogimpl

import (
	"io/fs"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func TestM2fBaseOpeLogs(t *testing.T) {
	lgr := common.DbgLogger()
	lgr.Debug("TestM2fBaseOpeLogs: start")

	std := t.TempDir()
	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := makeTestM2fManager(path.Join(ltd, "m2f.opl"), 0, true, std, ttd)
	require.NoError(t, err)

	olm.PutLogicalEntry("", opelog.NewLogicalEntry())
	require.NoError(t, olm.Sync())
	olm.PutLogicalEntry("a", opelog.NewLogicalEntry())
	require.NoError(t, olm.Sync())

	olm2, err := MakeM2fManager(path.Join(ltd, "m2f.opl"), 0)
	require.NoError(t, err)
	require.Error(t, olm2.Open(false))
	require.NoError(t, olm2.Open(true))
	require.NoError(t, olm.Close())
	require.NoError(t, olm2.Close())
	lgr.Debug("TestM2fBaseOpeLogs: end")
}

func TestM2fWriteOpeLogs(t *testing.T) {
	lgr := common.DbgLogger()
	lgr.Debug("TestM2fWriteOpeLogs: start")

	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 250, 25000, 2, 2025, false, 0))
	lgr.Debug("TestM2fWriteOpeLogs: FileTreeGenerated")

	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := makeTestM2fManager(path.Join(ltd, "m2f.opl"), 0, true, std, ttd)
	require.NoError(t, err)
	require.Contains(t, olm.Olm().Sessions, "ds")
	sTs := olm.Olm().Sessions["ds"].ReferenceTime

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

	olm2, err := MakeM2fManager(path.Join(ltd, "m2f.opl"), 0)
	require.NoError(t, err)
	require.Error(t, olm2.Open(false))
	require.NoError(t, olm2.Open(true))
	require.NoError(t, olm.Close())
	require.NoError(t, olm2.Close())
	lgr.Debug("TestM2fWriteOpeLogs: end")
}
