package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
)

const testHost = "localhost"

type opeServer struct {
	opegrpc.UnimplementedOpeServer
}

func (s *opeServer) Ready(context.Context, *opegrpc.Empty) (*opegrpc.Bool, error) {
	return &opegrpc.Bool{Value: true}, nil
}
func (s *opeServer) Version(context.Context, *opegrpc.Empty) (*opegrpc.Value, error) {
	return &opegrpc.Value{Value: "0.1"}, nil
}
func (s *opeServer) Shutdown(context.Context, *opegrpc.Value) (*opegrpc.Bool, error) {
	return &opegrpc.Bool{Value: true}, nil
}

func RunGrpcTestServer() (int, context.CancelFunc, error) {
	_, cCancel := context.WithCancel(context.Background())
	var (
		err  error
		port int
		opts []grpc.ServerOption
	)
	defer func() {
		if err != nil {
			cCancel()
		}
	}()
	if port, err = common.GetFreePort(); err != nil {
		return port, cCancel, err
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", testHost, port))
	grpcServer := grpc.NewServer(opts...)
	opegrpc.RegisterOpeServer(grpcServer, &opeServer{})
	go grpcServer.Serve(lis)
	cancel := func() {
		cCancel()
		grpcServer.GracefulStop()
	}
	return port, cancel, nil
}
