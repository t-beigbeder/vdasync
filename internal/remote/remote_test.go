package remote

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRunOpeDssaServer(t *testing.T) {
	td := t.TempDir()
	t.Chdir(td)
	common.WriteFile(t.Name()+".txt", []byte(t.Name()+"\n"))
	port, cFunc, err := RunOpeDssaServer(context.Background(), testHost, 0, nil, localfiles.MakeLocalFilesDssa(), nil)
	require.Nil(t, err)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	cli, conn, err := NewOpeDssaClient(fmt.Sprintf("%s:%d", testHost, port), opts...)
	require.Nil(t, err)
	defer conn.Close()

	rr, err := cli.Ready(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.True(t, rr.Value)
	rv, err := cli.Version(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.Equal(t, "0.1", rv.Value)

	wd, err := os.Getwd()
	require.Nil(t, err)
	rl, err := cli.List(context.Background(), &dssagrpc.Path{Path: wd})
	require.Nil(t, err)
	require.Equal(t, 1, len(rl.Entries))
	require.False(t, rl.Entries[0].IsDir)
	_, name := path.Split(rl.Entries[0].Path)
	require.Equal(t, t.Name()+".txt", name)

	cFunc()
}
