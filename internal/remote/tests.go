package remote

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const testHost = "localhost"

func doRunGrpcTestServer(tToListen time.Duration) (int, context.CancelFunc, error) {
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
	grpcServer := grpc.NewServer(opts...)
	go func() {
		if tToListen != 0 {
			time.Sleep(tToListen)
		}
		lis, lErr := net.Listen("tcp", fmt.Sprintf("%s:%d", testHost, port))
		if lErr != nil {
			return
		}
		opegrpc.RegisterOpeServer(grpcServer, &opeServer{grpcServer: grpcServer})
		dssagrpc.RegisterDataStorageSystemServer(grpcServer, &localFilesServer{grpcServer: grpcServer})
		grpcServer.Serve(lis)
	}()
	cancel := func() {
		cCancel()
		grpcServer.Stop()
	}
	return port, cancel, nil
}

func RunGrpcTestServer() (int, context.CancelFunc, error) {
	return doRunGrpcTestServer(0)
}

func checkLocalServerReadiness(port int) (
	cli OpeDssaClient, err error,
) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	return CheckServerReadiness(fmt.Sprintf("%s:%d", testHost, port), opts...)
}

func doGrpcGetTestClient(serverTToListen time.Duration, retryCount int, retryDelay time.Duration) (
	OpeDssaClient, context.CancelFunc, error,
) {
	var (
		cancel context.CancelFunc
		err    error
		cli    OpeDssaClient
	)
	port, cancel, err := doRunGrpcTestServer(serverTToListen)
	if err != nil {
		return nil, nil, fmt.Errorf("doGrpcGetTestClient: doRunGrpcTestServer failed %v", err)
	}
	for count := 0; count < retryCount; count++ {
		cli, err = checkLocalServerReadiness(port)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(retryDelay))
		retryDelay *= 2
	}
	return cli, cancel, nil
}

func GrpcGetTestClient() (
	OpeDssaClient, context.CancelFunc, error,
) {
	return doGrpcGetTestClient(0, 3, 20*time.Millisecond)
}
