package encrypted

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"github.com/t-beigbeder/vdasync/opegrpc"
)

const testHost = "localhost"

func TestProxy(t *testing.T) {
	lgr := common.GetLogger()
	td := t.TempDir()
	rec1, id1, err := common.AgeNewKeyPair()
	require.NoError(t, err)
	dss, err := MakeProxyDssa(lgr, td, []string{id1})
	require.NoError(t, err)
	require.Error(t, dss.NewSession())
	rec2, id2, err := common.AgeNewKeyPair()
	require.NoError(t, err)
	idss, err := common.AgeEncryptMsg([]byte(id2), rec1)
	require.NoError(t, err)
	require.NoError(t, dss.GetValueSetCb()(KeyIds, idss))
	require.NoError(t, dss.GetValueSetCb()(KeyRecs, []byte(rec2)))
	require.NoError(t, dss.GetValueSetCb()(KeyOpen, nil))
	require.NoError(t, dss.NewSession())
	require.NoError(t, dss.Mkdir(&dssa.DataEntry{Path: "/d1"}))
	require.NoError(t, dss.EndSession())
	require.NoError(t, dss.NewSession())
	de, err := dss.Stat("/d1")
	require.NoError(t, err)
	require.Equal(t, "/d1", de.Path)
	require.NoError(t, dss.EndSession())
	require.NoError(t, dss.GetValueSetCb()(KeyClose, nil))
	require.Error(t, dss.NewSession())
}

func TestRunOpeDssaTrustedServer(t *testing.T) {
	lgr := common.GetLogger()
	td := t.TempDir()
	rec1, id1, err := common.AgeNewKeyPair()
	require.NoError(t, err)
	dss, err := MakeProxyDssa(lgr, td, []string{id1})
	require.NoError(t, err)
	bgCtx := context.Background()
	port, cFunc, err := remote.RunOpeDssaServer(lgr, bgCtx,
		testHost, 0, nil, dss, nil, dss.GetValueSetCb())
	require.NoError(t, err)

	cli, conn, err := remote.NewOpeDssaClient(fmt.Sprintf("%s:%d", testHost, port), nil)
	require.NoError(t, err)
	defer conn.Close()

	_, err = cli.NewSession(bgCtx, &dssagrpc.Empty{})
	require.Error(t, err)

	rec2, id2, err := common.AgeNewKeyPair()
	require.NoError(t, err)
	idss, err := common.AgeEncryptMsg([]byte(id2), rec1)
	require.NoError(t, err)

	_, err = cli.SetValue(bgCtx, &opegrpc.KeyVal{Key: KeyIds, Val: idss})
	require.NoError(t, err)
	_, err = cli.SetValue(bgCtx, &opegrpc.KeyVal{Key: KeyRecs, Val: []byte(rec2)})
	require.NoError(t, err)
	_, err = cli.SetValue(bgCtx, &opegrpc.KeyVal{Key: KeyOpen})
	require.NoError(t, err)

	_, err = cli.NewSession(bgCtx, &dssagrpc.Empty{})
	require.NoError(t, err)
	_, err = cli.Mkdir(bgCtx, common.DssDte2GrpcDte(&dssa.DataEntry{Path: "/d1"}))
	require.NoError(t, err)
	_, err = cli.EndSession(bgCtx, &dssagrpc.Empty{})
	require.NoError(t, err)

	_, err = cli.NewSession(bgCtx, &dssagrpc.Empty{})
	require.NoError(t, err)
	gde, err := cli.Stat(bgCtx, &dssagrpc.Path{Path: "/d1"})
	require.NoError(t, err)
	de := common.GrpcDte2DssDte(gde)
	require.Equal(t, "/d1", de.Path)

	_, err = cli.SetValue(bgCtx, &opegrpc.KeyVal{Key: KeyClose})
	require.NoError(t, err)
	_, err = cli.NewSession(bgCtx, &dssagrpc.Empty{})
	require.Error(t, err)

	cFunc()
	time.Sleep(50 * time.Millisecond)
}
