package remote

import (
	"context"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
	des, err := cli.List(ctx, &dssagrpc.Path{Path: strings.Split(path.Dir(ft), "/")})
	require.Nil(t, err)
	require.Equal(t, 1, len(des.Entries))
}
