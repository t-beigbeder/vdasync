package sftpc

import (
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func wtf(ds dssa.Dssa, path_ string) error {
	wc, err := ds.GetWriteCloser(path_)
	if err != nil {
		return err
	}
	if _, err = wc.Write([]byte(path_ + "\n")); err != nil {
		return err
	}
	if err = wc.Close(); err != nil {
		return err
	}
	return nil
}

func TestSftpStuff(t *testing.T) {
	SkipIf(t)
	ds := GetSftpDss(t)
	require.NoError(t, Cleanup(ds))
	des, err := ds.List("/")
	require.NoError(t, err)
	require.Zero(t, len(des))
}

func TestBasicDirsAndFiles(t *testing.T) {
	SkipIf(t)
	ds := GetSftpDss(t)
	require.NoError(t, Cleanup(ds))
	des, err := ds.List("/")
	require.NoError(t, err)
	require.Zero(t, len(des))
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d1", IsDir: true}))
	des, err = ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 1, len(des))
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d2", IsDir: true}))
	require.NoError(t, wtf(ds, "/d1/f1.txt"))
	require.NoError(t, wtf(ds, "/d2/f2.txt"))
	require.NoError(t, wtf(ds, "/d2/f3.txt"))
	require.NoError(t, wtf(ds, "/f0.txt"))
	ls, err := ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	ls, err = ds.List("/d2")
	require.NoError(t, err)
	require.Equal(t, 2, len(ls))
	rc, err := ds.GetReadCloser("d1/f1.txt")
	require.NoError(t, err)
	b, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, "/d1/f1.txt\n", string(b))
	require.Error(t, ds.Rm("/d1"))
	require.NoError(t, ds.Rm("/d1/f1.txt"))
	require.NoError(t, ds.Rm("/d1"))
	de, err := ds.Stat("/d2/f2.txt")
	require.NoError(t, err)
	require.Equal(t, "/d2/f2.txt", de.Path)
	var de2 dssa.DataEntry
	de2 = *de
	de2.Mtime = de.Mtime - 600
	de2.GroupRights.Write = false
	require.NoError(t, ds.SetStat(&de2, false, false))
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d3", IsDir: true}))
	require.NoError(t, ds.Symlink("/d2/f3.txt", "/d3/f4.sl"))
	de3, err := ds.Stat("/d3/f4.sl")
	require.NoError(t, err)
	require.True(t, de3.IsSymLink)
	require.Equal(t, "/d2/f3.txt", de3.SymLinkTarget)
	des, err = ds.List("/d3")
	require.NoError(t, err)
	require.Equal(t, 1, len(des))
	require.True(t, des[0].NoLStat)
	require.Equal(t, "/d3/f4.sl", des[0].Path)
}

func TestConcurrency(t *testing.T) {
	SkipIf(t)
	ds := GetSftpDss(t)
	require.NoError(t, Cleanup(ds))
	var wg sync.WaitGroup
	lgr := common.DbgLogger()
	lgr.Debug("start")
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(count int, lgr *slog.Logger) {
			lgr.Debug("List")
			ds.List("/")
			lgr.Debug("Mkdir")
			ds.Mkdir(&dssa.DataEntry{Path: fmt.Sprintf("/d%d", count)})
			lgr.Debug("List again")
			des, err := ds.List("/")
			lgr.Debug("Done", "des", len(des), "err", err)
			wg.Done()
		}(i, lgr.With("concurrent", i))
	}
	lgr.Debug("wait")
	wg.Wait()
	lgr.Debug("done")
}
