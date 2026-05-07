package remote

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGrpcGetTestClientBase(t *testing.T) {
	td := t.TempDir()
	t.Chdir(td)
	common.WriteFile(t.Name()+".txt", []byte(t.Name()+"\n"))

	cli, cFunc, err := GrpcGetTestClient()
	require.Nil(t, err)

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

func TestRunGrpcTestServerFailSlowListen(t *testing.T) {
	port, cFunc, err := doRunGrpcTestServer(250 * time.Millisecond)
	require.Nil(t, err)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", testHost, port), opts...)
	require.Nil(t, err)
	defer conn.Close()
	client := opegrpc.NewOpeClient(conn)
	_, err = client.Ready(context.Background(), &opegrpc.Empty{})
	require.NotNil(t, err)
	cFunc()
}

func TestGrpcGetTestClientWaitSlowStart(t *testing.T) {
	td := t.TempDir()
	t.Chdir(td)
	common.WriteFile(t.Name()+".txt", []byte(t.Name()+"\n"))

	cli, cFunc, err := doGrpcGetTestClient(250*time.Millisecond, 5, 20*time.Millisecond)
	require.Nil(t, err)

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

func TestRunGrpcTestServerShutdown(t *testing.T) {
	cli, _, err := GrpcGetTestClient()
	require.Nil(t, err)

	rr, err := cli.Ready(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.True(t, rr.Value)

	rs, err := cli.Shutdown(context.Background(), &opegrpc.Value{Value: "10ms"})
	require.Nil(t, err)
	require.True(t, rs.Value)
}
