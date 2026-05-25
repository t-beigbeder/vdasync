package grpcclient

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/remote"
)

func TestFunctions(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.Nil(t, err)
	defer cFunc()
	dgc := MakeGrpcClient(common.GetLogger(), context.Background(), cli)
	des, err := dgc.List(path.Dir(ft))
	require.Nil(t, err)
	require.Equal(t, 1, len(des))

	de, err := dgc.Stat(ft)
	require.Nil(t, err)
	require.Nil(t, common.Lutimes(ft, de.Mtime-600)) // grpc server runs locally
	de2, err := dgc.Stat(ft)
	require.Nil(t, err)
	require.Equal(t, de.Mtime-600, de2.Mtime)

	de2.Mtime = de.Mtime
	de2.GroupRights = dssa.Rights{}
	de2.OtherRights = dssa.Rights{}
	err = dgc.SetStat(de2, false, false)
	require.Nil(t, err)
	de3, err := dgc.Stat(ft)
	require.Nil(t, err)
	require.Equal(t, de.Mtime, de3.Mtime)

	lt := path.Join(t.TempDir(), "TestFileFunctions.symlink")
	err = dgc.Symlink(ft, lt)
	require.Nil(t, err)
	lde, err := dgc.Stat(lt)
	require.Nil(t, err)
	require.True(t, lde.IsSymLink)
	require.Equal(t, ft, lde.SymLinkTarget)

	net := path.Join(t.TempDir(), "TestFileFunctionsNotYet.dat")
	de4, err := dgc.Stat(net)
	require.NotNil(t, err)
	require.NotNil(t, de4)
	require.True(t, de4.ErrNotExist)
	require.Equal(t, err, de4.Error)

	td := path.Join(t.TempDir(), "TestFileFunctionsMkdir")
	err = dgc.Mkdir(&dssa.DataEntry{Path: td, UserRights: dssa.Rights{Read: true, Write: true, Execute: true}})
	require.Nil(t, err)

	err = dgc.Rm(de3.Path)
	require.Nil(t, err)
	de3nd, err := dgc.Stat(de3.Path)
	require.NotNil(t, err)
	require.NotNil(t, de3nd)
	require.True(t, de3nd.ErrNotExist)

	err = dgc.Rm(td)
	require.Nil(t, err)
	dednd, err := dgc.Stat(td)
	require.NotNil(t, err)
	require.NotNil(t, dednd)
	require.True(t, dednd.ErrNotExist)

}

func TestWriter(t *testing.T) {
	tds := t.TempDir()
	tdt := t.TempDir()
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.Nil(t, err)
	defer cFunc()
	dgc := MakeGrpcClient(common.GetLogger(), context.Background(), cli)
	for ix, size := range []int64{0, 1023, 32*1024 - 1, 32 * 1024, 32*1024*1024 - 1, 32 * 1024 * 1024} {
		fn := fmt.Sprintf("TestWriter%d.dat", ix)
		fts := path.Join(tds, fn)
		ftd := path.Join(tdt, fn)

		err := common.MakeTestFile(fts, int(size))
		require.Nil(t, err)
		rdr, err := os.Open(fts)
		require.Nil(t, err)
		defer rdr.Close()

		wc, err := dgc.GetWriteCloser(ftd)
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
	tdd := t.TempDir()
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.Nil(t, err)
	defer cFunc()
	dgc := MakeGrpcClient(common.GetLogger(), context.Background(), cli)
	for ix, size := range []int64{1023, 32*1024 - 1, 32 * 1024, 32*1024*1024 - 1, 32 * 1024 * 1024} {
		for jx, wBufSize := range []int{1021, 32*1024 - 1, 32 * 1024, 32*1024 + 1, 32 * 1024 * 1024, 32*1024*1024 + 1} {
			fn := fmt.Sprintf("TestReader2-%d-%d.dat", ix, jx)
			fts := path.Join(tds, fn)
			ftd := path.Join(tdd, fn)

			err := common.MakeTestFile(fts, int(size)) // test server runs on localhost
			require.Nil(t, err)

			wrr, err := os.Create(ftd)
			require.Nil(t, err)
			defer wrr.Close()

			rc, err := dgc.GetReadCloser(fts)
			require.Nil(t, err)
			defer rc.Close()

			buffer := make([]byte, wBufSize)
			var lw int64
			for {
				n, err := rc.Read(buffer)
				if err != nil && err != io.EOF {
					break
				}
				wrr.Write(buffer[0:n])
				lw += int64(n)
				if err == io.EOF {
					err = nil
					break
				}
			}
			require.Nil(t, err)
			require.Equal(t, size, lw)
			wrr.Close()
			rc.Close()

			shs, err := common.FileSha256(fts)
			require.Nil(t, err)
			shd, err := common.FileSha256(ftd)
			require.Nil(t, err)
			require.Equal(t, shs, shd)
		}
	}
}
