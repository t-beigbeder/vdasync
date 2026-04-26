package remote

import (
	"context"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"google.golang.org/grpc"
)

type dssaImpl struct {
	dssagrpc.UnimplementedDataStorageSystemServer
	grpcServer *grpc.Server
	dssa_      dssa.Dssa
}

// List implements [dssagrpc.DataStorageSystemServer].
func (s *dssaImpl) List(ctx context.Context, dpath *dssagrpc.Path) (*dssagrpc.DataEntries, error) {
	ddtes, err := s.dssa_.List(dpath.Path)
	if err != nil {
		return nil, err
	}
	gdtes := dssagrpc.DataEntries{}
	for _, ddte := range ddtes {
		gdtes.Entries = append(gdtes.Entries, &dssagrpc.DataEntry{
			IsDir: ddte.IsDir,
			Path:  &dssagrpc.Path{Path: ddte.Path},
		})
	}
	return &gdtes, nil
}
