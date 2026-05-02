package remote

import (
	"context"
	"fmt"
	"net"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
)

type OpeDssaClient interface {
	opegrpc.OpeClient
	dssagrpc.DataStorageSystemClient
}

type opeDssaClient struct {
	opegrpc.OpeClient
	dssagrpc.DataStorageSystemClient
}

func NewOpeDssaClient(target string, opts ...grpc.DialOption) (OpeDssaClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, nil, err
	}
	oc := opegrpc.NewOpeClient(conn)
	dc := dssagrpc.NewDataStorageSystemClient(conn)
	return opeDssaClient{OpeClient: oc, DataStorageSystemClient: dc}, conn, nil
}

func CheckServerReadiness(target string, opts ...grpc.DialOption) (
	OpeDssaClient, error,
) {
	cli, conn, err := NewOpeDssaClient(target, opts...)
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

func RunOpeDssaServer(
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
	dssagrpc.RegisterDataStorageSystemServer(
		grpcServer,
		&dssaImpl{grpcServer: grpcServer, dssa_: dssa_},
	)
	go grpcServer.Serve(lis)
	cancel := func() {
		cCancel()
		grpcServer.Stop()
	}
	return port, cancel, nil
}
