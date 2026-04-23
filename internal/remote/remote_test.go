package remote

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRunOpeDssaServer(t *testing.T) {
	td := t.TempDir()
	t.Chdir(td)
	common.WriteFile(t.Name()+".txt", []byte(t.Name()+"\n"))
	port, cFunc, err := RunOpeDssaServer(context.Background(), testHost, 0, nil, NewLocalFilesServer, nil)
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

	rl, err := cli.List(context.Background(), &dssagrpc.Path{Path: "."})
	require.Nil(t, err)
	require.Equal(t, 1, len(rl.Entries))
	require.False(t, rl.Entries[0].IsDir)
	require.Equal(t, t.Name()+".txt", rl.Entries[0].Name)

	cFunc()
}
