package remote

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRunGrpcTestServerBase(t *testing.T) {
	td := t.TempDir()
	t.Chdir(td)
	common.WriteFile(t.Name()+".txt", []byte(t.Name()+"\n"))
	port, cFunc, err := RunGrpcTestServer()
	require.Nil(t, err)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", testHost, port), opts...)
	require.Nil(t, err)
	defer conn.Close()
	opeCli := opegrpc.NewOpeClient(conn)
	rr, err := opeCli.Ready(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.True(t, rr.Value)
	rv, err := opeCli.Version(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.Equal(t, "0.1", rv.Value)

	dsCli := dssagrpc.NewDataStorageSystemClient(conn)
	rl, err := dsCli.List(context.Background(), &dssagrpc.Path{Path: "."})
	require.Nil(t, err)
	require.Equal(t, 1, len(rl.Entries))
	require.False(t, rl.Entries[0].IsDir)
	require.Equal(t, t.Name()+".txt", rl.Entries[0].Name)
	cFunc()
}

func TestRunGrpcTestServerFailSlowListen(t *testing.T) {
	port, cFunc, err := doRunGrpcTestServer(250 * time.Millisecond)
	require.Nil(t, err)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", testHost, port), opts...)
	require.Nil(t, err)
	defer conn.Close()
	client := opegrpc.NewOpeClient(conn)
	_, err = client.Ready(context.Background(), &opegrpc.Empty{})
	require.NotNil(t, err)
	cFunc()
}

func TestRunGrpcTestServerShutdown(t *testing.T) {
	port, _, err := RunGrpcTestServer()
	require.Nil(t, err)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", testHost, port), opts...)
	require.Nil(t, err)
	defer conn.Close()
	client := opegrpc.NewOpeClient(conn)
	rr, err := client.Ready(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.True(t, rr.Value)
	rs, err := client.Shutdown(context.Background(), &opegrpc.Value{Value: "10ms"})
	require.Nil(t, err)
	require.True(t, rs.Value)
}

func TestGrpcGetTestClientBase(t *testing.T) {
	td := t.TempDir()
	t.Chdir(td)
	common.WriteFile(t.Name()+".txt", []byte(t.Name()+"\n"))

	opeCli, dsCli, cFunc, err := GrpcGetTestClient()
	require.Nil(t, err)

	rr, err := opeCli.Ready(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.True(t, rr.Value)
	rv, err := opeCli.Version(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.Equal(t, "0.1", rv.Value)

	rl, err := dsCli.List(context.Background(), &dssagrpc.Path{Path: "."})
	require.Nil(t, err)
	require.Equal(t, 1, len(rl.Entries))
	require.False(t, rl.Entries[0].IsDir)
	require.Equal(t, t.Name()+".txt", rl.Entries[0].Name)
	cFunc()

}

func TestGrpcGetTestClientWaitSlowStart(t *testing.T) {
	td := t.TempDir()
	t.Chdir(td)
	common.WriteFile(t.Name()+".txt", []byte(t.Name()+"\n"))

	opeCli, dsCli, cFunc, err := doGrpcGetTestClient(250*time.Millisecond, 5, 20*time.Millisecond)
	require.Nil(t, err)

	rr, err := opeCli.Ready(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.True(t, rr.Value)
	rv, err := opeCli.Version(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.Equal(t, "0.1", rv.Value)

	rl, err := dsCli.List(context.Background(), &dssagrpc.Path{Path: "."})
	require.Nil(t, err)
	require.Equal(t, 1, len(rl.Entries))
	require.False(t, rl.Entries[0].IsDir)
	require.Equal(t, t.Name()+".txt", rl.Entries[0].Name)
	cFunc()

}
