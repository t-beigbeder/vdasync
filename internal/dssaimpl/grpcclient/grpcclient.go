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
	gc.lgr.Info("grpcClient.EndSession")
	_, err := gc.client.EndSession(gc.ctx, &dssagrpc.Empty{})
	return err
}

// NewSession implements [dssa.Dssa].
func (gc *grpcClient) NewSession() error {
	gc.lgr.Info("grpcClient.NewSession")
	_, err := gc.client.NewSession(gc.ctx, &dssagrpc.Empty{})
	return err
}

// List implements [dssa.Dssa].
func (gc *grpcClient) List(path_ string) ([]*dssa.DataEntry, error) {
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
	_, err := gc.client.Mkdir(gc.ctx, common.DssDte2GrpcDte(de))
	return err
}

// Stat implements [dssa.Dssa].
func (gc *grpcClient) Stat(path_ string) (*dssa.DataEntry, error) {
	gd, err := gc.client.Stat(gc.ctx, &dssagrpc.Path{Path: path_})
	if err != nil {
		return nil, err
	}
	de := common.GrpcDte2DssDte(gd)
	return de, de.Error
}

// SetStat implements [dssa.Dssa].
func (gc *grpcClient) SetStat(ssde *dssa.DataEntry, noPerm, noMtime bool) error {
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
	return &grpcReader{gc: gc, path_: path_}, nil
}

// GetWriter implements [dssa.Dssa].
func (gc *grpcClient) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	return &grpcWriter{path_: path_, gc: gc}, nil
}

// Symlink implements [dssa.Dssa].
func (gc *grpcClient) Rm(path_ string) error {
	_, err := gc.client.Rm(gc.ctx, &dssagrpc.Path{Path: path_})
	return err
}

// Symlink implements [dssa.Dssa].
func (gc *grpcClient) Symlink(old string, new_ string) error {
	_, err := gc.client.Symlink(gc.ctx, &dssagrpc.OldNewPaths{Old: old, New_: new_})
	return err
}

func MakeGrpcClient(lgr *slog.Logger, ctx context.Context, client remote.OpeDssaClient) dssa.Dssa {
	return &grpcClient{lgr, ctx, client}
}
