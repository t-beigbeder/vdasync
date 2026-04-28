package grpcclient

import (
	"io"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

type grpcClient struct {
}

// List implements [dssa.Dssa].
func (d *grpcClient) List(path_ dssa.Path) ([]*dssa.DataEntry, error) {
	panic("")
}

// Stat implements [dssa.Dssa].
func (d *grpcClient) Stat(path_ dssa.Path) (*dssa.DataEntry, error) {
	panic("")
}

// SetStat implements [dssa.Dssa].
func (d *grpcClient) SetStat(de *dssa.DataEntry) error {
	panic("")
}

// GetReader implements [dssa.Dssa].
func (d *grpcClient) GetReadCloser(path_ dssa.Path) (io.ReadCloser, error) {
	panic("")
}

// GetWriter implements [dssa.Dssa].
func (d *grpcClient) GetWriteCloser(path_ dssa.Path) (io.WriteCloser, error) {
	panic("")
}

func MakeGrpcClient() dssa.Dssa {
	return &grpcClient{}
}
