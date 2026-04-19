package remote

import (
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
)

type OpeDssaClient interface {
	opegrpc.OpeClient
	dssagrpc.DataStorageSystemClient
}

func NewOpeDssaClient(target string, opts ...grpc.DialOption) (OpeDssaClient, error) {
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}
	oc := opegrpc.NewOpeClient(conn)
	dc := dssagrpc.NewDataStorageSystemClient(conn)
	_, _ = oc, dc
	return nil, err // FIXME
}
