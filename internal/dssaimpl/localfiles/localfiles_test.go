package localfiles

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

func TestFileFunctions(t *testing.T) {
	td1 := t.TempDir()
	ft := path.Join(td1, "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))
	lfd := MakeLocalFilesDssa()
	de, err := lfd.Stat(strings.Split(ft, "/"))
	require.Nil(t, err)
	require.Nil(t, common.Lutimes(ft, de.Mtime-600))
	de2, err := lfd.Stat(strings.Split(ft, "/"))
	require.Nil(t, err)
	require.Equal(t, de.Mtime-600, de2.Mtime)
	de2.Mtime = de.Mtime
	de2.GroupRights = dssa.Rights{}
	de2.OtherRights = dssa.Rights{}
	err = lfd.SetStat(de2)
	require.Nil(t, err)
	de3, err := lfd.Stat(strings.Split(ft, "/"))
	require.Nil(t, err)
	require.Equal(t, de.Mtime, de3.Mtime)

	lt := path.Join(t.TempDir(), "TestFileFunctions.symlink")
	err = os.Symlink(ft, lt)
	require.Nil(t, err)
	ltde, err := lfd.Stat(strings.Split(lt, "/"))
	require.Nil(t, err)
	require.True(t, ltde.IsSymLink)
	require.Equal(t, ft, ltde.SymLinkTarget)

	ldt := path.Join(t.TempDir(), "TestFileFunctions.lnsd")
	err = os.Symlink(td1, ldt)
	require.Nil(t, err)
	ldtde, err := lfd.Stat(strings.Split(ldt, "/"))
	require.Nil(t, err)
	require.True(t, ldtde.IsSymLink)
	require.False(t, ldtde.IsDir)
	require.Equal(t, td1, ldtde.SymLinkTarget)

	tt := path.Join(t.TempDir(), "TestFileFunctionsCopied.dat")
	rc, err := lfd.GetReadCloser(strings.Split(ft, "/"))
	require.Nil(t, err)
	defer rc.Close()
	wc, err := lfd.GetWriteCloser(strings.Split(tt, "/"))
	require.Nil(t, err)
	defer wc.Close()
	lw, err := io.Copy(wc, rc)
	require.Nil(t, err)
	require.Equal(t, de.Size, lw)

	nef := path.Join(t.TempDir(), "TestFileFunctionsNotYet.dat")
	de4, err := lfd.Stat(common.OsPath2DssPath(nef))
	require.NotNil(t, err)
	require.NotNil(t, de4)
	require.True(t, de4.ErrNotExist)
	require.Equal(t, err, de4.Error)

	td := path.Join(t.TempDir(), "TestFileFunctionsNotYetNewDir")
	err = lfd.Mkdir(&dssa.DataEntry{Path: common.OsPath2DssPath(td), UserRights: dssa.Rights{Read: true, Execute: true, Write: true}})
	require.Nil(t, err)
}

func TestFileGetPut(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestFileGetPut.dat")
	err := common.MakeTestFile(ft, 32*1024*1024)
	require.Nil(t, err)

	lfd := MakeLocalFilesDssa()
	de, err := lfd.Stat(strings.Split(ft, "/"))
	require.Nil(t, err)
	require.Equal(t, 32*1024*1024, int(de.Size))

	tt := path.Join(t.TempDir(), "TestFileGetPutCopied.dat")
	rc, err := lfd.GetReadCloser(strings.Split(ft, "/"))
	require.Nil(t, err)
	defer rc.Close()
	wc, err := lfd.GetWriteCloser(strings.Split(tt, "/"))
	require.Nil(t, err)
	defer wc.Close()
	written := 0
	buf := make([]byte, 31999)
	for {
		nr, err := rc.Read(buf)
		if err != nil {
			break
		}
		nw, err := wc.Write(buf[0:nr])
		if err != nil {
			break
		}
		written += nw
	}
	require.Nil(t, err)
	require.Equal(t, de.Size, int64(written))
}
