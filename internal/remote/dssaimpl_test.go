package remote

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

func TestFileFunctions(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))
	cli, cFunc, err := GrpcGetTestClient()
	require.Nil(t, err)
	defer cFunc()
	ctx := context.Background()
	des, err := cli.List(ctx, common.OsPath2GrpcPath(path.Dir(ft)))
	require.Nil(t, err)
	require.Equal(t, 1, len(des.Entries))
	gdte, err := cli.Stat(ctx, common.OsPath2GrpcPath(ft))
	require.Nil(t, err)
	ddte := common.GrpcDte2DssDte(gdte)
	require.Nil(t, common.Lutimes(ft, ddte.Mtime-600)) // grpc server runs locally
	gdte2, err := cli.Stat(ctx, common.OsPath2GrpcPath(ft))
	require.Nil(t, err)
	ddte2 := common.GrpcDte2DssDte(gdte2)
	require.Equal(t, ddte.Mtime-600, ddte2.Mtime)
	gdte2.Mtime = ddte.Mtime
	gdte2.GroupRights = common.DssRights2GrpcRights(dssa.Rights{})
	gdte2.OtherRights = common.DssRights2GrpcRights(dssa.Rights{})
	_, err = cli.SetStat(ctx, &dssagrpc.SetStatDataEntry{DataEntry: gdte2, NoPerm: false, NoMtime: false})
	require.Nil(t, err)
	gdte3, err := cli.Stat(ctx, common.OsPath2GrpcPath(ft))
	require.Nil(t, err)
	ddte3 := common.GrpcDte2DssDte(gdte3)
	require.Equal(t, ddte.Mtime, ddte3.Mtime)
	lt := path.Join(t.TempDir(), "TestFileFunctions.symlink")
	err = os.Symlink(ft, lt) // grpc server runs locally
	require.Nil(t, err)
	gldte, err := cli.Stat(ctx, common.OsPath2GrpcPath(lt))
	require.Nil(t, err)
	dldte := common.GrpcDte2DssDte(gldte)
	require.True(t, dldte.IsSymLink)
	require.Equal(t, ft, dldte.SymLinkTarget)
}
