package dssaimpl

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/grpcclient"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/proxy"
	"github.com/t-beigbeder/vdasync/internal/remote"
)

func TestAll(t *testing.T) {
	t.Skip("wip") // TODO: what was it?
	tds, err := NewTestDss(&TestDssOptions{
		Lgr:         common.DbgLogger(),
		Kind:        "sftp",
		PluginLevel: "DEBUG",
	})
	require.NoError(t, err)
	require.NoError(t, tds.Close())
}

func TestProxyWithGrpc(t *testing.T) {
	lgr := common.GetLogger()
	// lgr = common.DbgLogger()

	ft := path.Join(t.TempDir(), "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))
	cli, cFunc, err := remote.GrpcGetTestClient(nil, nil)
	require.Nil(t, err)
	defer cFunc()
	dgc := grpcclient.MakeGrpcClient(common.GetLogger(), context.Background(), cli)
	des, err := dgc.List(path.Dir(ft))
	require.Nil(t, err)
	require.Equal(t, 1, len(des))

	pDss := proxy.MakeRateLimitedDss(lgr.With("source", true), dgc, 2, 6e6)
	des, err = pDss.List(path.Dir(ft))
	require.Nil(t, err)
	require.Equal(t, 1, len(des))
	require.True(t, pDss.IsGrpcClient())
	gc, _ := pDss.(grpcclient.GrpcClient)
	rc := gc.GetClient()
	require.NotNil(t, rc)

	pDss = proxy.MakeRateLimitedDss(lgr.With("source", true), localfiles.MakeLocalFilesDssa(), 2, 6e6)
	des, err = pDss.List(path.Dir(ft))
	require.Nil(t, err)
	require.Equal(t, 1, len(des))
	require.False(t, pDss.IsGrpcClient())

}
