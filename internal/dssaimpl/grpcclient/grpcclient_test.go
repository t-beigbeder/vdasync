package grpcclient

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/remote"
)

func TestFunctions(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))
	cli, cFunc, err := remote.GrpcGetTestClient()
	require.Nil(t, err)
	defer cFunc()
	dgc := MakeGrpcClient(context.Background(), cli)
	des, err := dgc.List(common.OsPath2DssPath(path.Dir(ft)))
	require.Nil(t, err)
	require.Equal(t, 1, len(des))

	de, err := dgc.Stat(common.OsPath2DssPath(ft))
	require.Nil(t, err)
	require.Nil(t, common.Lutimes(ft, de.Mtime-600)) // grpc server runs locally
	de2, err := dgc.Stat(common.OsPath2DssPath(ft))
	require.Nil(t, err)
	require.Equal(t, de.Mtime-600, de2.Mtime)

	de2.Mtime = de.Mtime
	de2.GroupRights = dssa.Rights{}
	de2.OtherRights = dssa.Rights{}
	err = dgc.SetStat(de2)
	require.Nil(t, err)
	de3, err := dgc.Stat(common.OsPath2DssPath(ft))
	require.Nil(t, err)
	require.Equal(t, de.Mtime, de3.Mtime)

	lt := path.Join(t.TempDir(), "TestFileFunctions.symlink")
	err = os.Symlink(ft, lt) // grpc server runs locally
	require.Nil(t, err)
	lde, err := dgc.Stat(common.OsPath2DssPath(lt))
	require.Nil(t, err)
	require.True(t, lde.IsSymLink)
	require.Equal(t, ft, lde.SymLinkTarget)
}

func TestWriter(t *testing.T) {
	tds := t.TempDir()
	tdt := t.TempDir()
	cli, cFunc, err := remote.GrpcGetTestClient()
	require.Nil(t, err)
	defer cFunc()
	dgc := MakeGrpcClient(context.Background(), cli)
	for ix, size := range []int64{1023, 32*1024 - 1, 32*1024*1024 - 1, 32 * 1024 * 1024} {
		fn := fmt.Sprintf("TestWriter%d.dat", ix)
		fts := path.Join(tds, fn)
		ftd := path.Join(tdt, fn)

		err := common.MakeTestFile(fts, int(size))
		require.Nil(t, err)
		rdr, err := os.Open(fts)
		require.Nil(t, err)
		defer rdr.Close()

		wc, err := dgc.GetWriteCloser(common.OsPath2DssPath(ftd))
		require.Nil(t, err)
		defer wc.Close()

		lw, err := io.Copy(wc, rdr)
		require.Nil(t, err)
		require.Equal(t, size, lw)
		rdr.Close()
		wc.Close()

		fi, err := os.Stat(ftd)
		require.Nil(t, err)
		require.Equal(t, size, fi.Size())
	}
}

func TestReader(t *testing.T) {
	tds := t.TempDir()
	tdt := t.TempDir()
	cli, cFunc, err := remote.GrpcGetTestClient()
	require.Nil(t, err)
	defer cFunc()
	dgc := MakeGrpcClient(context.Background(), cli)
	for ix, size := range []int64{1023, 32*1024 - 1, 32*1024*1024 - 1, 32 * 1024 * 1024} {
		fn := fmt.Sprintf("TestReader%d.dat", ix)
		fts := path.Join(tds, fn)
		ftd := path.Join(tdt, fn)

		err := common.MakeTestFile(fts, int(size)) // test server runs on localhost
		require.Nil(t, err)

		wrr, err := os.Create(ftd)
		require.Nil(t, err)
		defer wrr.Close()

		rc, err := dgc.GetReadCloser(common.OsPath2DssPath(fts))
		require.Nil(t, err)
		defer rc.Close()

		lw, err := io.Copy(wrr, rc)
		require.Nil(t, err)
		require.Equal(t, size, lw)
		wrr.Close()
		rc.Close()
	}

}
