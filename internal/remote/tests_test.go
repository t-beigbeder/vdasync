package remote

import (
	"context"
	ctls "crypto/tls"
	"fmt"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/tls"
	"github.com/t-beigbeder/vdasync/opegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

func TestRunGrpcTestServerSelfSigned(t *testing.T) {
	td := t.TempDir()
	cf := path.Join(td, "self-cert.pem")
	kf := path.Join(td, "self-key.pem")
	err := tls.SelfSignedFiles("localhost", cf, kf)
	require.Nil(t, err)
	creds, err := credentials.NewServerTLSFromFile(cf, kf)
	require.Nil(t, err)

	address := fmt.Sprintf("%s:%d", "localhost", 9443)

	go func() {
		s := grpc.NewServer(grpc.Creds(creds))
		lis, err := net.Listen("tcp", address)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Listen failed %v\n", err)
			time.Sleep(10 * time.Second)
		}
		s.Serve(lis)
		fmt.Fprintf(os.Stderr, "Serve done\n")
	}()
	config := &ctls.Config{
		InsecureSkipVerify: true,
	}
	time.Sleep(1 * time.Second)
	cli, _, err := NewOpeDssaClient(address, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	require.Nil(t, err)
	rs, err := cli.Shutdown(context.Background(), &opegrpc.Value{Value: "500ms"})
	require.Nil(t, err)
	_ = rs
	time.Sleep(1 * time.Second)
}
