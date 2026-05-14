package remote

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"runtime"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OpeDssaClient interface {
	opegrpc.OpeClient
	dssagrpc.DataStorageSystemClient
}

type opeDssaClient struct {
	opegrpc.OpeClient
	dssagrpc.DataStorageSystemClient
}

func GetDefaultCopt(copt grpc.DialOption) grpc.DialOption {
	if copt == nil {
		copt = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	return copt
}

func NewOpeDssaClient(target string, copt grpc.DialOption) (OpeDssaClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(target, GetDefaultCopt(copt))
	if err != nil {
		return nil, nil, err
	}
	oc := opegrpc.NewOpeClient(conn)
	dc := dssagrpc.NewDataStorageSystemClient(conn)
	return opeDssaClient{OpeClient: oc, DataStorageSystemClient: dc}, conn, nil
}

func CheckServerReadiness(target string, copt grpc.DialOption) (
	OpeDssaClient, error,
) {
	cli, conn, err := NewOpeDssaClient(target, GetDefaultCopt(copt))
	if err != nil {
		return nil, err
	}
	_, err = cli.Ready(context.Background(), &opegrpc.Empty{})
	if err != nil {
		conn.Close()
		return nil, err
	}
	return cli, err
}

func getStat(lgr *slog.Logger, callStat chan string) {
	count := 0
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	lgr.Info("RunOpeDssaServer: starting", "HeapInuse", m.HeapInuse/1024, "HeapAlloc", m.HeapAlloc/1024, "StackInuse", m.StackInuse/1024)
	statMap := make(map[string]int)
	for stat := range callStat {
		count++
		statMap[stat]++
		if count%1000 == 0 {
			lgr.Info("RunOpeDssaServer: processed...", "count", count,
				"HeapInuse", m.HeapInuse/1024, "HeapAlloc", m.HeapAlloc/1024, "StackInuse", m.StackInuse/1024,
				"statMap", statMap)
		}
		_ = stat
	}
	lgr.Info("RunOpeDssaServer: done", "count", count,
		"HeapInuse", m.HeapInuse/1024, "HeapAlloc", m.HeapAlloc/1024, "StackInuse", m.StackInuse/1024,
		"statMap", statMap)
}

func RunOpeDssaServer(
	lgr *slog.Logger,
	ctx context.Context,
	host string,
	port int,
	opts []grpc.ServerOption,
	dssa_ dssa.Dssa,
	shutdownCb func(),
) (
	int, context.CancelFunc, error,
) {
	_, cCancel := context.WithCancel(context.Background())
	var (
		err error
	)
	defer func() {
		if err != nil {
			cCancel()
		}
	}()
	if port == 0 {
		if port, err = common.GetFreePort(); err != nil {
			return port, cCancel, err
		}
	}
	grpcServer := grpc.NewServer(opts...)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return port, cCancel, err
	}
	opegrpc.RegisterOpeServer(grpcServer, &opeServer{grpcServer: grpcServer, shutdownCb: shutdownCb})

	callStats := make(chan string)
	go getStat(lgr, callStats)
	dssagrpc.RegisterDataStorageSystemServer(
		grpcServer,
		&dssaImpl{grpcServer: grpcServer, dssa_: dssa_, callStats: callStats},
	)
	go grpcServer.Serve(lis)
	cancel := func() {
		cCancel()
		grpcServer.Stop()
		close(callStats)
	}
	return port, cancel, nil
}
