package grpcclient

import (
	"context"
	"io"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/remote"
)

type grpcClient struct {
	ctx    context.Context
	client remote.OpeDssaClient
}

// List implements [dssa.Dssa].
func (gc *grpcClient) List(path_ dssa.Path) ([]*dssa.DataEntry, error) {
	gds, err := gc.client.List(gc.ctx, common.DssPath2GrpcPath(path_))
	if err != nil {
		return nil, err
	}
	dds := []*dssa.DataEntry{}
	for _, gd := range gds.Entries {
		dds = append(dds, common.GrpcDte2DssDte(gd))
	}
	return dds, nil
}

// Stat implements [dssa.Dssa].
func (gc *grpcClient) Stat(path_ dssa.Path) (*dssa.DataEntry, error) {
	gd, err := gc.client.Stat(gc.ctx, common.DssPath2GrpcPath(path_))
	if err != nil {
		return nil, err
	}
	de := common.GrpcDte2DssDte(gd)
	return de, de.Error
}

// SetStat implements [dssa.Dssa].
func (gc *grpcClient) SetStat(de *dssa.DataEntry) error {
	_, err := gc.client.SetStat(gc.ctx, common.DssDte2GrpcDte(de))
	return err
}

// GetReader implements [dssa.Dssa].
func (gc *grpcClient) GetReadCloser(path_ dssa.Path) (io.ReadCloser, error) {
	return &grpcReader{gc: gc, path_: path_}, nil
}

// GetWriter implements [dssa.Dssa].
func (gc *grpcClient) GetWriteCloser(path_ dssa.Path) (io.WriteCloser, error) {
	return &grpcWriter{path_: path_, gc: gc}, nil
}

// Symlink implements [dssa.Dssa].
func (gc *grpcClient) Symlink(old dssa.Path, new_ dssa.Path) error {
	_, err := gc.client.Symlink(gc.ctx, &dssagrpc.OldNewPaths{Old: old, New_: new_})
	return err
}

func MakeGrpcClient(ctx context.Context, client remote.OpeDssaClient) dssa.Dssa {
	return &grpcClient{ctx, client}
}
