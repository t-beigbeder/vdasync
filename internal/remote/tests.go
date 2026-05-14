package remote

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/opegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const testHost = "localhost"

func doRunGrpcTestServer(tToListen time.Duration, opt ...grpc.ServerOption) (int, context.CancelFunc, error) {
	_, cCancel := context.WithCancel(context.Background())
	var (
		err  error
		port int
	)
	defer func() {
		if err != nil {
			cCancel()
		}
	}()
	if port, err = common.GetFreePort(); err != nil {
		return port, cCancel, err
	}
	grpcServer := grpc.NewServer(opt...)
	callStats := make(chan string)

	go func() {
		if tToListen != 0 {
			time.Sleep(tToListen)
		}
		lis, lErr := net.Listen("tcp", fmt.Sprintf("%s:%d", testHost, port))
		if lErr != nil {
			return
		}
		opegrpc.RegisterOpeServer(grpcServer, &opeServer{grpcServer: grpcServer})

		go getStat(common.GetLogger(), callStats)
		dssagrpc.RegisterDataStorageSystemServer(
			grpcServer,
			&dssaImpl{grpcServer: grpcServer, dssa_: localfiles.MakeLocalFilesDssa(), callStats: callStats},
		)
		grpcServer.Serve(lis)
		common.GetLogger().Error("doRunGrpcTestServer: stopped serving")
	}()
	cancel := func() {
		cCancel()
		grpcServer.Stop()
		close(callStats)
	}
	return port, cancel, nil
}

func RunGrpcTestServer(opt ...grpc.ServerOption) (int, context.CancelFunc, error) {
	return doRunGrpcTestServer(0, opt...)
}

func checkLocalServerReadiness(port int) (
	cli OpeDssaClient, err error,
) {
	config := &tls.Config{
		InsecureSkipVerify: false,
	}
	tcskip := credentials.NewTLS(config)
	tcinsec := insecure.NewCredentials()
	_, _ = tcinsec, tcskip
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tcinsec)}
	return CheckServerReadiness(fmt.Sprintf("%s:%d", testHost, port), opts...)
}

func doGrpcGetTestClient(serverTToListen time.Duration, retryCount int, retryDelay time.Duration, opt ...grpc.ServerOption) (
	OpeDssaClient, context.CancelFunc, error,
) {
	var (
		cancel context.CancelFunc
		err    error
		cli    OpeDssaClient
	)
	port, cancel, err := doRunGrpcTestServer(serverTToListen, opt...)
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
	if err != nil {
		return nil, nil, err
	}
	return cli, cancel, nil
}

func GrpcGetTestClient(opt ...grpc.ServerOption) (
	OpeDssaClient, context.CancelFunc, error,
) {
	return doGrpcGetTestClient(0, 3, 20*time.Millisecond, opt...)
}
