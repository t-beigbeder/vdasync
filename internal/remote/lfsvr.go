package remote

import (
	"context"
	"os"

	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"google.golang.org/grpc"
)

type localFilesServer struct {
	dssagrpc.UnimplementedDataStorageSystemServer
	grpcServer *grpc.Server
}

// List implements [dssagrpc.DataStorageSystemServer].
func (s *localFilesServer) List(ctx context.Context, p *dssagrpc.Path) (*dssagrpc.DataEntries, error) {
	des, err := os.ReadDir(p.Path)
	if err != nil {
		return nil, err
	}
	var dtes dssagrpc.DataEntries
	for _, de := range des {
		dtes.Entries = append(dtes.Entries, &dssagrpc.DataEntry{Name: de.Name(), IsDir: de.IsDir()})
	}
	return &dtes, nil
}

func NewLocalFilesServer(grpcServer *grpc.Server) dssagrpc.DataStorageSystemServer {
	return &localFilesServer{grpcServer: grpcServer}
}
