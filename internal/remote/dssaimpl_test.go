package remote

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

func TestFileFunctions(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))
	cli, cFunc, err := GrpcGetTestClient()
	require.Nil(t, err)
	defer cFunc()
	ctx := context.Background()
	des, err := cli.List(ctx, os2gp(path.Dir(ft)))
	require.Nil(t, err)
	require.Equal(t, 1, len(des.Entries))
	gdte, err := cli.Stat(ctx, os2gp(ft))
	require.Nil(t, err)
	ddte := gdte2ddte(gdte)
	require.Nil(t, common.Lutimes(ft, ddte.Mtime-600)) // grpc server runs locally
	gdte2, err := cli.Stat(ctx, os2gp(ft))
	require.Nil(t, err)
	ddte2 := gdte2ddte(gdte2)
	require.Equal(t, ddte.Mtime-600, ddte2.Mtime)
	gdte2.Mtime = ddte.Mtime
	gdte2.GroupRights = drts2grts(dssa.Rights{})
	gdte2.OtherRights = drts2grts(dssa.Rights{})
	_, err = cli.SetStat(ctx, gdte2)
	require.Nil(t, err)
	gdte3, err := cli.Stat(ctx, os2gp(ft))
	require.Nil(t, err)
	ddte3 := gdte2ddte(gdte3)
	require.Equal(t, ddte.Mtime, ddte3.Mtime)
}
