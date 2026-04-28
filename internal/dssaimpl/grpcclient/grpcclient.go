package grpcclient

import (
	"context"
	"io"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/remote"
)

type grpcClient struct {
	ctx    context.Context
	client remote.OpeDssaClient
}

// List implements [dssa.Dssa].
func (d *grpcClient) List(path_ dssa.Path) ([]*dssa.DataEntry, error) {
	gds, err := d.client.List(d.ctx, common.DssPath2GrpcPath(path_))
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
func (d *grpcClient) Stat(path_ dssa.Path) (*dssa.DataEntry, error) {
	gd, err := d.client.Stat(d.ctx, common.DssPath2GrpcPath(path_))
	if err != nil {
		return nil, err
	}
	return common.GrpcDte2DssDte(gd), nil
}

// SetStat implements [dssa.Dssa].
func (d *grpcClient) SetStat(de *dssa.DataEntry) error {
	_, err := d.client.SetStat(d.ctx, common.DssDte2GrpcDte(de))
	return err
}

// GetReader implements [dssa.Dssa].
func (d *grpcClient) GetReadCloser(path_ dssa.Path) (io.ReadCloser, error) {
	panic("")
	// returns a reader that implements client receiving server streaming
	// each reader read() provides available buffered data with
	// optional recv() to get more
}

// GetWriter implements [dssa.Dssa].
func (d *grpcClient) GetWriteCloser(path_ dssa.Path) (io.WriteCloser, error) {
	panic("")
	// returns a writer that implements client streaming
	// each writer write() translates to a grpc call to send() after initial put()
}

func MakeGrpcClient(ctx context.Context, client remote.OpeDssaClient) dssa.Dssa {
	return &grpcClient{ctx, client}
}
