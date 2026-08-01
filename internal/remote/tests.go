package remote

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/opegrpc"
	"google.golang.org/grpc"
)

const testHost = "localhost"

type TestProxyDssa interface {
	dssa.Dssa
	GetValueSetCb() func(string, []byte) error
}

func doRunGrpcTestServer(tToListen time.Duration, tpDss TestProxyDssa, opt ...grpc.ServerOption) (int, context.CancelFunc, error) {
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
		var tDss dssa.Dssa
		var vsCb func(key string, value []byte) error
		if tpDss == nil {
			tDss = localfiles.MakeLocalFilesDssa()
		} else {
			tDss = tpDss
			vsCb = tpDss.GetValueSetCb()
		}
		opegrpc.RegisterOpeServer(grpcServer,
			&opeServer{grpcServer: grpcServer, valueSetCb: vsCb})

		lgr := common.GetNullLogger()
		go getStat(lgr, callStats)
		dssagrpc.RegisterDataStorageSystemServer(
			grpcServer,
			&dssaImpl{lgr: lgr, grpcServer: grpcServer, dssa_: tDss, callStats: callStats},
		)
		grpcServer.Serve(lis)
		lgr.Error("doRunGrpcTestServer: stopped serving")
	}()
	cancel := func() {
		cCancel()
		grpcServer.Stop()
		close(callStats)
	}
	return port, cancel, nil
}

func checkLocalServerReadiness(port int, copt grpc.DialOption) (
	cli OpeDssaClient, err error,
) {
	return CheckServerReadiness(fmt.Sprintf("%s:%d", testHost, port), CoptOrDefault(copt))
}

func doGrpcGetTestClient(serverTToListen time.Duration, retryCount int, retryDelay time.Duration,
	tDss TestProxyDssa,
	copt grpc.DialOption, sopt ...grpc.ServerOption) (
	OpeDssaClient, context.CancelFunc, error,
) {
	var (
		cancel context.CancelFunc
		err    error
		cli    OpeDssaClient
	)
	port, cancel, err := doRunGrpcTestServer(serverTToListen, tDss, sopt...)
	if err != nil {
		return nil, nil, fmt.Errorf("doGrpcGetTestClient: doRunGrpcTestServer failed %v", err)
	}
	for count := 0; count < retryCount; count++ {
		cli, err = checkLocalServerReadiness(port, copt)
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

func GrpcGetTestClient(tDss TestProxyDssa, copt grpc.DialOption, sopt ...grpc.ServerOption) (
	OpeDssaClient, context.CancelFunc, error,
) {
	return doGrpcGetTestClient(0, 3, 20*time.Millisecond, tDss, copt, sopt...)
}
