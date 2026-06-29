package grpcclient

import (
	"context"
	"io"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/remote"
)

type grpcClient struct {
	lgr    *slog.Logger
	ctx    context.Context
	client remote.OpeDssaClient
}

// EndSession implements [dssa.Dssa].
func (gc *grpcClient) EndSession() error {
	gc.lgr.Debug("grpcClient.EndSession")
	_, err := gc.client.EndSession(gc.ctx, &dssagrpc.Empty{})
	return err
}

// NewSession implements [dssa.Dssa].
func (gc *grpcClient) NewSession() error {
	gc.lgr.Debug("grpcClient.NewSession")
	_, err := gc.client.NewSession(gc.ctx, &dssagrpc.Empty{})
	return err
}

// List implements [dssa.Dssa].
func (gc *grpcClient) List(path_ string) ([]*dssa.DataEntry, error) {
	gc.lgr.Debug("grpcClient.List", "path", path_)
	gds, err := gc.client.List(gc.ctx, &dssagrpc.Path{Path: path_})
	if err != nil {
		return nil, err
	}
	dds := []*dssa.DataEntry{}
	for _, gd := range gds.Entries {
		dds = append(dds, common.GrpcDte2DssDte(gd))
	}
	return dds, nil
}

// Mkdir implements [dssa.Dssa].
func (gc *grpcClient) Mkdir(de *dssa.DataEntry) error {
	gc.lgr.Debug("grpcClient.Mkdir", "de", de.Path)
	_, err := gc.client.Mkdir(gc.ctx, common.DssDte2GrpcDte(de))
	return err
}

// Stat implements [dssa.Dssa].
func (gc *grpcClient) Stat(path_ string) (*dssa.DataEntry, error) {
	gc.lgr.Debug("grpcClient.Stat", "de", path_)
	gd, err := gc.client.Stat(gc.ctx, &dssagrpc.Path{Path: path_})
	if err != nil {
		return nil, err
	}
	de := common.GrpcDte2DssDte(gd)
	return de, de.Error
}

// SetStat implements [dssa.Dssa].
func (gc *grpcClient) SetStat(ssde *dssa.DataEntry, noPerm, noMtime bool) error {
	gc.lgr.Debug("grpcClient.SetStat", "de", ssde.Path)
	_, err := gc.client.SetStat(gc.ctx,
		&dssagrpc.SetStatDataEntry{
			DataEntry: common.DssDte2GrpcDte(ssde),
			NoPerm:    noPerm,
			NoMtime:   noMtime,
		})
	return err
}

// GetReader implements [dssa.Dssa].
func (gc *grpcClient) GetReadCloser(path_ string) (io.ReadCloser, error) {
	gc.lgr.Debug("grpcClient.GetReadCloser", "path", path_)
	return &grpcReader{lgr: gc.lgr, gc: gc, path_: path_}, nil
}

// GetWriter implements [dssa.Dssa].
func (gc *grpcClient) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	gc.lgr.Debug("grpcClient.GetWriteCloser", "path", path_)
	return &grpcWriter{lgr: gc.lgr, path_: path_, gc: gc}, nil
}

// Symlink implements [dssa.Dssa].
func (gc *grpcClient) Rm(path_ string) error {
	gc.lgr.Debug("grpcClient.Rm", "path", path_)
	_, err := gc.client.Rm(gc.ctx, &dssagrpc.Path{Path: path_})
	return err
}

// Symlink implements [dssa.Dssa].
func (gc *grpcClient) Symlink(old string, new_ string) error {
	gc.lgr.Debug("grpcClient.Symlink", "path", new_)
	_, err := gc.client.Symlink(gc.ctx, &dssagrpc.OldNewPaths{Old: old, New_: new_})
	return err
}

func MakeGrpcClient(lgr *slog.Logger, ctx context.Context, client remote.OpeDssaClient) dssa.Dssa {
	return &grpcClient{lgr, ctx, client}
}
