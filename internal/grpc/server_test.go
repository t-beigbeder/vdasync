package grpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	gorpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServerBasic(t *testing.T) {
	port, cFunc, err := RunGrpcTestServer()
	require.Nil(t, err)
	var opts []gorpc.DialOption
	opts = append(opts, gorpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := gorpc.NewClient(fmt.Sprintf("%s:%d", testHost, port), opts...)
	require.Nil(t, err)
	defer conn.Close()
	client := opegrpc.NewOpeClient(conn)
	v, err := client.Version(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.Equal(t, "0.1", v.Value)
	cFunc()
}
