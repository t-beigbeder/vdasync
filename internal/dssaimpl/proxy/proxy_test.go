package proxy

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

func TestBasicProxy(t *testing.T) {
	lgr := common.GetLogger()
	udss := localfiles.MakeLocalFilesDssa()
	pdss := MakeProxyDssa(lgr, udss, nil)

	require.True(t, true)
	td1 := t.TempDir()
	ft := path.Join(td1, "TestBasicProxy.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))

	de, err := pdss.Stat(ft)
	require.Nil(t, err)
	require.Nil(t, common.Lutimes(ft, de.Mtime-600))
	de2, err := pdss.Stat(ft)
	require.Nil(t, err)
	require.Equal(t, de.Mtime-600, de2.Mtime)
	de2.Mtime = de.Mtime
	de2.GroupRights = dssa.Rights{}
	de2.OtherRights = dssa.Rights{}
	err = pdss.SetStat(de2, false, false)
	require.Nil(t, err)
	de3, err := pdss.Stat(ft)
	require.Nil(t, err)
	require.Equal(t, de.Mtime, de3.Mtime)

	lt := path.Join(t.TempDir(), "TestBasicProxy.symlink")
	err = os.Symlink(ft, lt)
	require.Nil(t, err)
	ltde, err := pdss.Stat(lt)
	require.Nil(t, err)
	require.True(t, ltde.IsSymLink)
	require.Equal(t, ft, ltde.SymLinkTarget)

	ldt := path.Join(t.TempDir(), "TestBasicProxy.lnsd")
	err = os.Symlink(td1, ldt)
	require.Nil(t, err)
	ldtde, err := pdss.Stat(ldt)
	require.Nil(t, err)
	require.True(t, ldtde.IsSymLink)
	require.False(t, ldtde.IsDir)
	require.Equal(t, td1, ldtde.SymLinkTarget)

	tt := path.Join(t.TempDir(), "TestBasicProxyCopied.dat")
	rc, err := pdss.GetReadCloser(ft)
	require.Nil(t, err)
	defer rc.Close()
	wc, err := pdss.GetWriteCloser(tt)
	require.Nil(t, err)
	defer wc.Close()
	lw, err := io.Copy(wc, rc)
	require.Nil(t, err)
	require.Equal(t, de.Size, lw)

	nef := path.Join(t.TempDir(), "TestBasicProxyNotYet.dat")
	de4, err := pdss.Stat(nef)
	require.NotNil(t, err)
	require.NotNil(t, de4)
	require.True(t, de4.ErrNotExist)
	require.Equal(t, err, de4.Error)

	td := path.Join(t.TempDir(), "TestBasicProxyNotYetNewDir")
	err = pdss.Mkdir(&dssa.DataEntry{Path: td, UserRights: dssa.Rights{Read: true, Execute: true, Write: true}})
	require.Nil(t, err)

	err = pdss.Rm(de3.Path)
	require.Nil(t, err)
	de3ne, err := pdss.Stat(de3.Path)
	require.NotNil(t, err)
	require.NotNil(t, de3ne)
	require.True(t, de3ne.ErrNotExist)

	err = pdss.Rm(td)
	require.Nil(t, err)
	dedne, err := pdss.Stat(td)
	require.NotNil(t, err)
	require.NotNil(t, dedne)
	require.True(t, dedne.ErrNotExist)

	ft = path.Join(td1, "TestBasicProxy.dat")
	require.Nil(t, common.WriteFile(ft, []byte("TestChecksum")))
	h1, err := pdss.Checksum("sha256", ft)
	require.Nil(t, err)
	require.Equal(t, "sha256:4b86be7f5fe5776cd535cdf1e81fdd77c204df48c751f61c121b3e72f6767e1e", h1)

	des, err := pdss.List(td1)
	require.NoError(t, err)
	_ = des
}
