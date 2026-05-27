package s3msts

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func wtf(ds S3DssaWithMsts, path_ string) error {
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

func TestBasicDirsAndFiles(t *testing.T) {
	SkipIf(t)
	ds := GetRepo(t)
	require.NotNil(t, ds)
	require.NoError(t, Cleanup(ds))
	require.NoError(t, ds.Msts().NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d1", IsDir: true}))
	require.NoError(t, ds.Msts().EndSession())
	require.NoError(t, ds.Msts().NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d2", IsDir: true}))
	require.NoError(t, ds.Msts().EndSession())
	require.NoError(t, ds.Msts().NewSession())
	require.NoError(t, wtf(ds, "/d1/f1.txt"))
	require.NoError(t, wtf(ds, "/d2/f2.txt"))
	require.NoError(t, wtf(ds, "/d2/f3.txt"))
	require.NoError(t, wtf(ds, "/f0.txt"))
	ls, err := ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	require.NoError(t, ds.Msts().EndSession())
	require.NoError(t, ds.Msts().NewSession())
	ls, err = ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	ls, err = ds.List("/d2")
	require.NoError(t, err)
	require.Equal(t, 2, len(ls))
}

func TestBisBasicDirsAndFiles(t *testing.T) {
	SkipIf(t)
	ds := GetRepo(t)
	require.NotNil(t, ds)
	require.NoError(t, Cleanup(ds))
	require.NoError(t, ds.NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d1", IsDir: true}))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d2", IsDir: true}))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	require.NoError(t, wtf(ds, "/d1/f1.txt"))
	require.NoError(t, wtf(ds, "/d2/f2.txt"))
	require.NoError(t, wtf(ds, "/d2/f3.txt"))
	require.NoError(t, wtf(ds, "/f0.txt"))
	ls, err := ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	ls, err = ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	ls, err = ds.List("/d2")
	require.NoError(t, err)
	require.Equal(t, 2, len(ls))
	de, err := ds.Stat("/d1/f1.txt")
	require.NoError(t, err)
	de2 := &dssa.DataEntry{}
	*de2 = *de
	de2.Mtime -= 600
	de2.GroupRights = dssa.Rights{}
	de2.OtherRights = dssa.Rights{}
	require.NoError(t, ds.SetStat(de2, false, false))
	require.NoError(t, ds.Symlink("/d1/f1.txt", "/d2/f4.symlink"))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	de3, err := ds.Stat("/d1/f1.txt")
	require.NoError(t, err)
	require.Equal(t, de2.Mtime, de3.Mtime)
	ls, err = ds.List("/d2")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	de4, err := ds.Stat("/d2/f4.symlink")
	require.NoError(t, err)
	require.True(t, de4.IsSymLink)
	require.Equal(t, "/d1/f1.txt", de4.SymLinkTarget)
	td := t.TempDir()
	ft := path.Join(td, "TestFileGetPut.dat")
	require.NoError(t, common.MakeTextTestFile(ft, 32*1024*1024))
	rc, err := os.Open(ft)
	require.NoError(t, err)
	defer rc.Close()
	wc, err := ds.GetWriteCloser("/d2/TestFileGetPut.dat")
	require.Nil(t, err)
	defer wc.Close()
	written := 0
	buf := make([]byte, 31999)
	for {
		nr, err := rc.Read(buf)
		if err != nil && err != io.EOF {
			break
		}
		nw, werr := wc.Write(buf[0:nr])
		if werr != nil {
			break
		}
		written += nw
		if err == io.EOF {
			break
		}
	}
	require.NoError(t, err)
	require.Equal(t, 32*1024*1024, written)
	require.NoError(t, wc.Close())
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	ls, err = ds.List("/d2")
	require.NoError(t, err)
	require.Equal(t, 4, len(ls))
	rc2, err := ds.GetReadCloser("/d2/TestFileGetPut.dat")
	require.Nil(t, err)
	ft2 := path.Join(td, "TestFileGetPutCopied.dat")
	wc2, err := os.Create(ft2)
	require.NoError(t, err)
	defer wc2.Close()
	written = 0
	buf = make([]byte, 31999)
	for {
		nr, err := rc2.Read(buf)
		if err != nil && err != io.EOF {
			break
		}
		nw, errw := wc2.Write(buf[0:nr])
		if errw != nil {
			break
		}
		written += nw
		if err == io.EOF {
			break
		}
	}
	require.NoError(t, err)
	require.Equal(t, 32*1024*1024, written)
	require.NoError(t, wc2.Close())
	require.NoError(t, rc2.Close())
	sh1, err := common.FileSha256(ft)
	require.NoError(t, err)
	sh2, err := common.FileSha256(ft2)
	require.NoError(t, err)
	require.Equal(t, sh1, sh2)
}

func TestBasicFilesRW(t *testing.T) {
	SkipIf(t)
	ds := GetRepo(t)
	require.NotNil(t, ds)
	require.NoError(t, Cleanup(ds))
	require.NoError(t, ds.NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d2", IsDir: true}))
	td := t.TempDir()
	ft := path.Join(td, "TestFileGetPut.dat")
	require.NoError(t, common.MakeTextTestFile(ft, 64*1024))
	rc, err := os.Open(ft)
	require.NoError(t, err)
	defer rc.Close()
	wc, err := ds.GetWriteCloser("/d2/TestFileGetPut.dat")
	require.Nil(t, err)
	defer wc.Close()
	written := 0
	buf := make([]byte, 31999)
	for {
		nr, err := rc.Read(buf)
		if err != nil && err != io.EOF {
			break
		}
		nw, werr := wc.Write(buf[0:nr])
		if werr != nil {
			break
		}
		written += nw
		if err == io.EOF {
			break
		}
	}
	require.NoError(t, err)
	require.Equal(t, 64*1024, written)
	require.NoError(t, wc.Close())
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	ls, err := ds.List("/d2")
	require.NoError(t, err)
	require.Equal(t, 1, len(ls))
	rc2, err := ds.GetReadCloser("/d2/TestFileGetPut.dat")
	require.Nil(t, err)
	ft2 := path.Join(td, "TestFileGetPutCopied.dat")
	wc2, err := os.Create(ft2)
	require.NoError(t, err)
	defer wc2.Close()
	written = 0
	buf = make([]byte, 31999)
	for {
		nr, err := rc2.Read(buf)
		if err != nil && err != io.EOF {
			break
		}
		nw, errw := wc2.Write(buf[0:nr])
		if errw != nil {
			break
		}
		written += nw
		if err == io.EOF {
			break
		}
	}
	require.NoError(t, err)
	require.Equal(t, 64*1024, written)
	require.NoError(t, wc2.Close())
	require.NoError(t, rc2.Close())
	sh1, err := common.FileSha256(ft)
	require.NoError(t, err)
	sh2, err := common.FileSha256(ft2)
	require.NoError(t, err)
	require.Equal(t, sh1, sh2)
}
