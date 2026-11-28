package opelogimpl

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
	"google.golang.org/protobuf/proto"
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
	ole, err := olm.GetLogicalEntry("c")
	require.NoError(t, err)
	require.Nil(t, ole)
	require.NoError(t, olm.PutLogicalEntry("a", opelog.NewLogicalEntry()))
	ole, err = olm.GetLogicalEntry("a")
	require.NoError(t, err)
	require.NotNil(t, ole)

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
	cliLgr, _ := common.CliLogger("TestPbfWalk", "DEBUG+2", "stderr")
	_ = cliLgr
	lgr = cliLgr
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
	lgr.Debug("TestPbfWalk: walked", "len(rps1)", len(rps1))

	btd := t.TempDir()
	require.NoError(t, olm.Open(true))
	require.NoError(t, olm.Backup(btd, 10000))
	require.NoError(t, olm.Close())

	rtd := t.TempDir()
	orlm, err := MakePbfManager(lgr, path.Join(rtd, "master.pbf"), 0)
	require.NoError(t, orlm.Restore(btd))
	require.NoError(t, orlm.Close())

	require.NoError(t, orlm.Open(true))
	total := 0
	number := 0
	orlm.Walk(func(relPath string, ole *opelog.LogicalEntry) error {
		_, _ = relPath, ole
		if ole == nil {
			lgr.Error("TestPbfWalk", "ole", "nil")
			return nil
		}
		gole := opelog.LogicalEntry2ProtoBuf(ole)
		bole, err := proto.Marshal(gole)
		if err != nil {
			lgr.Error("TestPbfWalk", "err", err)
			return nil
		}
		total += len(bole)
		number++
		if len(bole) > 2000 {
			lgr.Debug("TestPbfWalk", "relPath", relPath, "size", len(bole))
		}
		return nil
	}, false, false)
	require.NoError(t, orlm.Close())
	lgr.Info("TestPbfWalk", "stats", fmt.Sprintf("total %d - number %d - avr %d", total, number, total/number))

}

func TestOplSupport(t *testing.T) {
	t.Skip("just for support")
	lgr := common.GetLogger()
	// lgr = common.DbgLogger()
	cliLgr, _ := common.CliLogger("TestOplSupport", "DEBUG+2", "stderr")
	_ = cliLgr
	lgr = cliLgr
	lgr.Debug("TestPbfWalk: start")
	path_ := "/tmp/opd/mst.pbf"
	olm, err := MakePbfManager(lgr, path_, 0)
	require.NoError(t, err)
	require.NoError(t, olm.Open(true))
	total := 0
	number := 0
	olm.Walk(func(relPath string, ole *opelog.LogicalEntry) error {
		_, _ = relPath, ole
		if ole == nil {
			lgr.Error("TestOplSupport", "ole", "nil")
			return nil
		}
		gole := opelog.LogicalEntry2ProtoBuf(ole)
		bole, err := proto.Marshal(gole)
		if err != nil {
			lgr.Error("TestOplSupport", "err", err)
			return nil
		}
		total += len(bole)
		number++
		if strings.HasPrefix(relPath, "tmp/opd/") && len(relPath) == 10 {
			return nil
		}
		if strings.HasPrefix(relPath, ".git/objects/") && len(relPath) == 15 {
			return nil
		}
		if len(bole) > 2000 {
			lgr.Debug("TestOplSupport", "relPath", relPath, "size", len(bole))
		}
		return nil
	}, false, false)
	lgr.Info("TestOplSupport", "stats", fmt.Sprintf("total %d - number %d - avr %d", total, number, total/number))
	require.NoError(t, olm.Close())

	require.NoError(t, olm.Open(true))

	os.Mkdir("/local/tmp/opd_bck", 0750)
	require.NoError(t, olm.Backup("/local/tmp/opd_bck", 1000))
	require.NoError(t, olm.Close())

	ltd := t.TempDir()
	orlm, err := MakePbfManager(lgr, path.Join(ltd, "mst.pbf"), 0)
	require.NoError(t, orlm.Restore("/local/tmp/opd_bck"))

	require.NoError(t, orlm.Close())

	require.NoError(t, orlm.Open(true))
	total = 0
	number = 0
	orlm.Walk(func(relPath string, ole *opelog.LogicalEntry) error {
		_, _ = relPath, ole
		if ole == nil {
			lgr.Error("TestOplSupport", "ole", "nil")
			return nil
		}
		gole := opelog.LogicalEntry2ProtoBuf(ole)
		bole, err := proto.Marshal(gole)
		if err != nil {
			lgr.Error("TestOplSupport", "err", err)
			return nil
		}
		total += len(bole)
		number++
		if strings.HasPrefix(relPath, "tmp/opd/") && len(relPath) == 10 {
			return nil
		}
		if strings.HasPrefix(relPath, ".git/objects/") && len(relPath) == 15 {
			return nil
		}
		if len(bole) > 2000 {
			lgr.Debug("TestOplSupport", "relPath", relPath, "size", len(bole))
		}
		return nil
	}, false, false)
	require.NoError(t, orlm.Close())
	lgr.Info("TestOplSupport", "stats", fmt.Sprintf("total %d - number %d - avr %d", total, number, total/number))
}
