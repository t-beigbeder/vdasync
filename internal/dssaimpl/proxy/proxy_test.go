package proxy

import (
	"io"
	"log/slog"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

func doTestProxy(t *testing.T, lgr *slog.Logger, sDss, tDss dssa.Dssa, sHdl, tHdl ProxyHandler) {
	sPDss, tPDss := sDss, tDss
	if sHdl != nil {
		sPDss = MakeProxyDssa(lgr, sDss, sHdl)
	}
	if tHdl != nil {
		tPDss = MakeProxyDssa(lgr, tDss, tHdl)
	}

	td1 := t.TempDir()
	ft := path.Join(td1, "TestBasicProxy.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))

	de, err := sPDss.Stat(ft)
	require.Nil(t, err)
	require.Nil(t, common.Lutimes(ft, de.Mtime-600))
	de2, err := sPDss.Stat(ft)
	require.Nil(t, err)
	require.Equal(t, de.Mtime-600, de2.Mtime)
	de2.Mtime = de.Mtime
	de2.GroupRights = dssa.Rights{}
	de2.OtherRights = dssa.Rights{}
	err = sPDss.SetStat(de2, false, false)
	require.Nil(t, err)
	de3, err := sPDss.Stat(ft)
	require.Nil(t, err)
	require.Equal(t, de.Mtime, de3.Mtime)

	lt := path.Join(t.TempDir(), "TestBasicProxy.symlink")
	err = os.Symlink(ft, lt)
	require.Nil(t, err)
	ltde, err := sPDss.Stat(lt)
	require.Nil(t, err)
	require.True(t, ltde.IsSymLink)
	require.Equal(t, ft, ltde.SymLinkTarget)

	ldt := path.Join(t.TempDir(), "TestBasicProxy.lnsd")
	err = os.Symlink(td1, ldt)
	require.Nil(t, err)
	ldtde, err := sPDss.Stat(ldt)
	require.Nil(t, err)
	require.True(t, ldtde.IsSymLink)
	require.False(t, ldtde.IsDir)
	require.Equal(t, td1, ldtde.SymLinkTarget)

	tt := path.Join(t.TempDir(), "TestBasicProxyCopied.dat")
	rc, err := sPDss.GetReadCloser(ft)
	require.Nil(t, err)
	defer rc.Close()
	wc, err := tPDss.GetWriteCloser(tt)
	require.Nil(t, err)
	defer wc.Close()
	lw, err := io.Copy(wc, rc)
	require.Nil(t, err)
	require.Equal(t, de.Size, lw)

	nef := path.Join(t.TempDir(), "TestBasicProxyNotYet.dat")
	de4, err := sPDss.Stat(nef)
	require.NotNil(t, err)
	require.NotNil(t, de4)
	require.True(t, de4.ErrNotExist)
	require.Equal(t, err, de4.Error)

	td := path.Join(t.TempDir(), "TestBasicProxyNotYetNewDir")
	err = sPDss.Mkdir(&dssa.DataEntry{Path: td, UserRights: dssa.Rights{Read: true, Execute: true, Write: true}})
	require.Nil(t, err)

	err = sPDss.Rm(de3.Path)
	require.Nil(t, err)
	de3ne, err := sPDss.Stat(de3.Path)
	require.NotNil(t, err)
	require.NotNil(t, de3ne)
	require.True(t, de3ne.ErrNotExist)

	err = sPDss.Rm(td)
	require.Nil(t, err)
	dedne, err := sPDss.Stat(td)
	require.NotNil(t, err)
	require.NotNil(t, dedne)
	require.True(t, dedne.ErrNotExist)

	ft = path.Join(td1, "TestBasicProxy.dat")
	require.Nil(t, common.WriteFile(ft, []byte("TestChecksum")))
	h1, err := sPDss.Checksum("sha256", ft)
	require.Nil(t, err)
	require.Equal(t, "sha256:4b86be7f5fe5776cd535cdf1e81fdd77c204df48c751f61c121b3e72f6767e1e", h1)

	des, err := sPDss.List(td1)
	require.NoError(t, err)
	_ = des

	ft = path.Join(td1, "TestBasicProxyLarge.dat")
	common.MakeTestFile(ft, 1000000)
	h1, err = common.FileChecksum(ft, "sha256")
	require.NoError(t, err)
	tt = path.Join(t.TempDir(), "TestBasicProxyLargeCopied.dat")
	rc, err = sPDss.GetReadCloser(ft)
	require.Nil(t, err)
	defer rc.Close()
	wc, err = tPDss.GetWriteCloser(tt)
	require.Nil(t, err)
	defer wc.Close()
	lw, err = io.Copy(wc, rc)
	require.Nil(t, err)
	require.Equal(t, int64(1000000), lw)
	h2, err := common.FileChecksum(tt, "sha256")
	require.Nil(t, err)
	require.Equal(t, h1, h2)
}

func TestBasicProxy(t *testing.T) {
	lgr := common.GetLogger()
	// lgr = common.DbgLogger()
	udss := localfiles.MakeLocalFilesDssa()
	noh := &noOpHandler{lgr: lgr, underlying: udss}
	doTestProxy(t, lgr, udss, udss, noh, noh)
}

func TestRtLimBaseProxy(t *testing.T) {
	lgr := common.GetLogger()
	// lgr = common.DbgLogger()
	sDss := localfiles.MakeLocalFilesDssa()
	sRtlHdl := MakeRateLimHandler(lgr.With("isTarget", false), sDss, 3, 6e6)
	sRtlHdl.GetEbl().Record()
	tDss := localfiles.MakeLocalFilesDssa()
	tRtlHdl := MakeRateLimHandler(lgr.With("isTarget", true), tDss, 2, 6e6)
	tRtlHdl.GetEbl().Record()
	doTestProxy(t, lgr, sDss, tDss, sRtlHdl, tRtlHdl)
	nEv, nBytes, du, evr, bdw := sRtlHdl.GetEbl().Counts()
	lgr.Debug("test bandwidth src", "nev", nEv, "nbytes", nBytes, "duration", du.Seconds(), "evr", evr, "bdw", bdw)
	nEv, nBytes, du, evr, bdw = tRtlHdl.GetEbl().Counts()
	lgr.Debug("test bandwidth tgt", "nev", nEv, "nbytes", nBytes, "duration", du.Seconds(), "evr", evr, "bdw", bdw)
}

func TestRtLimDirectProxy(t *testing.T) {
	lgr := common.GetLogger()
	// lgr = common.DbgLogger()
	sDss := MakeRateLimitedDss(lgr.With("source", true), localfiles.MakeLocalFilesDssa(), 2, 6e6)
	tDss := MakeRateLimitedDss(lgr.With("target", true), localfiles.MakeLocalFilesDssa(), 2, 6e6)

	td1 := t.TempDir()
	ft := path.Join(td1, "TestBasicProxyLarge.dat")
	common.MakeTestFile(ft, 1000000)
	h1, err := common.FileChecksum(ft, "sha256")
	require.NoError(t, err)
	tt := path.Join(t.TempDir(), "TestBasicProxyLargeCopied.dat")
	rc, err := sDss.GetReadCloser(ft)
	require.Nil(t, err)
	defer rc.Close()
	wc, err := tDss.GetWriteCloser(tt)
	require.Nil(t, err)
	defer wc.Close()
	lw, err := io.Copy(wc, rc)
	require.Nil(t, err)
	require.Equal(t, int64(1000000), lw)
	h2, err := common.FileChecksum(tt, "sha256")
	require.Nil(t, err)
	require.Equal(t, h1, h2)

	nEv, nBytes, du, evr, bdw := sDss.GetEbl().Counts()
	lgr.Debug("test bandwidth src", "nev", nEv, "nbytes", nBytes, "duration", du.Seconds(), "evr", evr, "bdw", bdw)
	nEv, nBytes, du, evr, bdw = tDss.GetEbl().Counts()
	lgr.Debug("test bandwidth tgt", "nev", nEv, "nbytes", nBytes, "duration", du.Seconds(), "evr", evr, "bdw", bdw)

}

func TestIoErrDirectProxy(t *testing.T) {
	lgr := common.GetLogger()
	//lgr = common.DbgLogger()
	sDss := MakeIoErrDss(lgr.With("source", true), localfiles.MakeLocalFilesDssa(), nil)
	tDss := MakeIoErrDss(lgr.With("target", true), localfiles.MakeLocalFilesDssa(), nil)

	td1 := t.TempDir()
	ft := path.Join(td1, "TestBasicProxyLarge.dat")
	common.MakeTestFile(ft, 1000000)
	h1, err := common.FileChecksum(ft, "sha256")
	require.NoError(t, err)

	tt := path.Join(t.TempDir(), "TestBasicProxyLargeCopied.dat")
	rc, err := sDss.GetReadCloser(ft)
	require.Nil(t, err)
	defer rc.Close()
	wc, err := tDss.GetWriteCloser(tt)
	require.Nil(t, err)
	defer wc.Close()
	lw, err := io.Copy(wc, rc)
	require.Nil(t, err)
	require.Equal(t, int64(1000000), lw)
	h2, err := common.FileChecksum(tt, "sha256")
	require.Nil(t, err)
	require.Equal(t, h1, h2)

}

func TestRetryErrStandardProxy(t *testing.T) {
	lgr := common.GetLogger()
	// lgr = common.DbgLogger()
	sDss := MakeRetryErrDss(lgr.With("source", true), localfiles.MakeLocalFilesDssa(), nil)
	tDss := MakeRetryErrDss(lgr.With("target", true), localfiles.MakeLocalFilesDssa(), nil)

	doTestProxy(t, lgr, sDss, tDss, nil, nil)

}

func TestIoErrStandardProxy(t *testing.T) {
	lgr := common.GetLogger()
	// lgr = common.DbgLogger()
	opts := &IoErrOpts{
		ListRate: 100, MkdirRate: 100, StatRate: 100, SetStatRate: 100,
		GetReadCloserRate: 100, ReadRate: 200, RCloseRate: 100,
		GetWriteCloserRate: 100, WriteRate: 200, WCloseRate: 100,
		RmRate: 100, SymlinkRate: 100,
	}
	sIoeDss := MakeIoErrDss(lgr.With("source", true), localfiles.MakeLocalFilesDssa(), opts)
	tIoeDss := MakeIoErrDss(lgr.With("target", true), localfiles.MakeLocalFilesDssa(), opts)
	sDss := MakeRetryErrDss(lgr.With("source", true), sIoeDss, &RetryErrOpts{max: 3})
	tDss := MakeRetryErrDss(lgr.With("target", true), tIoeDss, &RetryErrOpts{max: 3})
	doTestProxy(t, lgr, sDss, tDss, nil, nil)

}
