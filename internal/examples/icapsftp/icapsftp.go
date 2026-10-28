package icapsftp

import (
	"io"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/sftpc"
)

type icapSftp struct {
	hasSession  bool
	crossOsSafe bool
}

// NewSession implements [dssa.Dssa].
func (i *icapSftp) NewSession() error {
	panic("unimplemented")
}

// EndSession implements [dssa.Dssa].
func (i *icapSftp) EndSession() error {
	panic("unimplemented")
}

// IsGrpcClient implements [dssa.Dssa].
func (i *icapSftp) IsGrpcClient() bool {
	panic("unimplemented")
}

// DssInfo implements [dssa.Dssa].
func (i *icapSftp) DssInfo() string {
	panic("unimplemented")
}

// List implements [dssa.Dssa].
func (i *icapSftp) List(string) ([]*dssa.DataEntry, error) {
	panic("unimplemented")
}

// Mkdir implements [dssa.Dssa].
func (i *icapSftp) Mkdir(*dssa.DataEntry) error {
	panic("unimplemented")
}

// Stat implements [dssa.Dssa].
func (i *icapSftp) Stat(string) (*dssa.DataEntry, error) {
	panic("unimplemented")
}

// SetStat implements [dssa.Dssa].
func (i *icapSftp) SetStat(_ *dssa.DataEntry, noPerm bool, noMtime bool) error {
	panic("unimplemented")
}

// Checksum implements [dssa.Dssa].
func (i *icapSftp) Checksum(algos string, path_ string) (string, error) {
	panic("unimplemented")
}

// GetReadCloser implements [dssa.Dssa].
func (i *icapSftp) GetReadCloser(string) (io.ReadCloser, error) {
	panic("unimplemented")
}

// GetWriteCloser implements [dssa.Dssa].
func (i *icapSftp) GetWriteCloser(string) (io.WriteCloser, error) {
	panic("unimplemented")
}

// Rm implements [dssa.Dssa].
func (i *icapSftp) Rm(string) error {
	panic("unimplemented")
}

// Symlink implements [dssa.Dssa].
func (i *icapSftp) Symlink(old string, new_ string) error {
	panic("unimplemented")
}

func MakeIcapSftpDssa(	user, address, identity, root string, concurrency int,
	factory sftpc.SftpClientFactory, knownHostsFile string) (dssa.Dssa, error) {
	return &icapSftp{}, nil
}
