package grpcclient

import (
	"context"
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

	// FIXME: continue
}
