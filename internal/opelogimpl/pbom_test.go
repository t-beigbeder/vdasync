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

func TestPbfBaseOpeLogs(t *testing.T) {
	lgr := common.GetLogger()
	lgr.Debug("TestPbfBaseOpeLogs: start")

	std := t.TempDir()
	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := makeTestPbfManager(lgr, path.Join(ltd, "master.pbf"), 0, std, ttd)
	require.NoError(t, err)

	require.NoError(t, olm.PutLogicalEntry("", opelog.NewLogicalEntry()))
	require.NoError(t, olm.Sync())
	require.NoError(t, olm.PutLogicalEntry("a", opelog.NewLogicalEntry()))
	require.NoError(t, olm.Sync())

	olm2, err := MakePbfManager(lgr, path.Join(ltd, "master.pbf"), 0)
	require.NoError(t, err)
	require.Error(t, olm2.Open(false))
	require.NoError(t, olm2.Open(true))
	require.NoError(t, olm.Close())
	require.NoError(t, olm2.Close())
	lgr.Debug("TestPbfBaseOpeLogs: end")
}

func TestPbfWriteOpeLogs(t *testing.T) {
	lgr := common.GetLogger()
	lgr.Debug("TestPbfWriteOpeLogs: start")

	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 250, 25000, 2, 2025, false, 0))
	lgr.Debug("TestPbfWriteOpeLogs: FileTreeGenerated")

	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := makeTestPbfManager(lgr, path.Join(ltd, "master.pbf"), 0, std, ttd)
	require.NoError(t, err)
	require.Contains(t, olm.GetOlm().Sessions, "ds")
	sTs := olm.GetOlm().Sessions["ds"].ReferenceTime

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

	olm2, err := MakePbfManager(lgr, path.Join(ltd, "master.pbf"), 0)
	require.NoError(t, err)
	require.Error(t, olm2.Open(false))
	require.NoError(t, olm2.Open(true))
	require.NoError(t, olm.Close())
	require.NoError(t, olm2.Close())
	lgr.Debug("TestPbfWriteOpeLogs: end")
}

func TestPbfWalk(t *testing.T) {
	lgr := common.GetLogger()
	lgr = common.DbgLogger()
	lgr.Debug("TestPbfWalk: start")

	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 250, 25000, 2, 2025, false, 0))
	lgr.Debug("TestPbfWalk: FileTreeGenerated")

	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := makeTestPbfManager(lgr, path.Join(ltd, "master.pbf"), 0, std, ttd)
	require.NoError(t, err)
	require.Contains(t, olm.GetOlm().Sessions, "ds")
	sTs := olm.GetOlm().Sessions["ds"].ReferenceTime

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
	var (
		rps1, rps2, rps3 []string
	)
	_, _, _ = rps1, rps2, rps3
	olm.Walk(func(relPath string, ole *opelog.LogicalEntry) error {
		rps1 = append(rps1, relPath)
		return nil
	}, false, false)
	olm.Walk(func(relPath string, ole *opelog.LogicalEntry) error {
		rps2 = append(rps2, relPath)
		return nil
	}, true, false)
	olm.Walk(func(relPath string, ole *opelog.LogicalEntry) error {
		rps3 = append(rps3, relPath)
		return nil
	}, true, true)
	require.Equal(t, len(rps1), len(rps2))
	require.Equal(t, len(rps1), len(rps3))
	require.NoError(t, olm.Close())
	lgr.Debug("TestPbfWalk: end")
}
