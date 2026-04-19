package remote

import (
	"context"
	"fmt"
	"net"

	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
)

const testHost = "localhost"

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
	dssagrpc.RegisterDataStorageSystemServer(grpcServer, &localFilesServer{})
	go grpcServer.Serve(lis)
	cancel := func() {
		cCancel()
		grpcServer.Stop()
	}
	return port, cancel, nil
}
