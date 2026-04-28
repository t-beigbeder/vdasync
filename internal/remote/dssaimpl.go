package remote

import (
	"context"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"google.golang.org/grpc"
)

type dssaImpl struct {
	dssagrpc.UnimplementedDataStorageSystemServer
	grpcServer *grpc.Server
	dssa_      dssa.Dssa
}

// List implements [dssagrpc.DataStorageSystemServer].
func (s *dssaImpl) List(ctx context.Context, gpath *dssagrpc.Path) (*dssagrpc.DataEntries, error) {
	ddtes, err := s.dssa_.List(gpath.Path)
	if err != nil {
		return nil, err
	}
	gdtes := dssagrpc.DataEntries{}
	for _, ddte := range ddtes {
		gdtes.Entries = append(gdtes.Entries, common.DssDte2GrpcDte(ddte))
	}
	return &gdtes, nil
}

func (s *dssaImpl) Stat(ctx context.Context, gpath *dssagrpc.Path) (*dssagrpc.DataEntry, error) {
	ddte, err := s.dssa_.Stat(gpath.Path)
	if err != nil {
		return nil, err
	}
	return common.DssDte2GrpcDte(ddte), nil
}

func (s *dssaImpl) SetStat(ctx context.Context, gdte *dssagrpc.DataEntry) (*dssagrpc.Empty, error) {
	if err := s.dssa_.SetStat(common.GrpcDte2DssDte(gdte)); err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}

func (s *dssaImpl) Put(grpc.ClientStreamingServer[dssagrpc.PushedBlock, dssagrpc.Length]) error {
	panic("")
	// client-side streaming
	// gets a writer to Dssa
	// while recv-ing on the stream, write to Dssa
}

func (s *dssaImpl) Get(*dssagrpc.Path, grpc.ServerStreamingServer[dssagrpc.PulledBlock]) error {
	panic("")
	// server side streaming
	// get a reader from Dssa
	// while read-ing on the Dssa, send to the stream
}
