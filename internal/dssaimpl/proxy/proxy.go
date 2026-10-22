package proxy

import (
	"io"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
)

type ProxyDssa interface {
	dssa.Dssa
	Underlying() dssa.Dssa
}

type ProxyHdlArgs struct {
	Path    string
	De      *dssa.DataEntry
	NoPerm  bool
	NoMtime bool
	Algos   string
	Old     string
	New     string
	RC      io.ReadCloser
	WC      io.WriteCloser
	Buf     []byte
}

type ProxyHdlRess struct {
	Err    error
	Des    []*dssa.DataEntry
	De     *dssa.DataEntry
	RC     io.ReadCloser
	WC     io.WriteCloser
	ChSums string
	NRdWr  int
}

type ProxyHandler interface {
	Handle(fcName string, args *ProxyHdlArgs) *ProxyHdlRess
}

type proxyDssa struct {
	lgr        *slog.Logger
	underlying dssa.Dssa
	handler    ProxyHandler
}

// NewSession implements [ProxyDssa].
func (pd *proxyDssa) NewSession() error {
	ress := pd.handler.Handle("NewSession", &ProxyHdlArgs{})
	return ress.Err
}

// EndSession implements [ProxyDssa].
func (pd *proxyDssa) EndSession() error {
	ress := pd.handler.Handle("EndSession", &ProxyHdlArgs{})
	return ress.Err
}

// IsGrpcClient implements [dssa.Dssa].
func (pd *proxyDssa) IsGrpcClient() bool { return pd.underlying.IsGrpcClient() }

// List implements [ProxyDssa].
func (pd *proxyDssa) List(path_ string) ([]*dssa.DataEntry, error) {
	ress := pd.handler.Handle("List", &ProxyHdlArgs{Path: path_})
	return ress.Des, ress.Err
}

// Mkdir implements [ProxyDssa].
func (pd *proxyDssa) Mkdir(de *dssa.DataEntry) error {
	ress := pd.handler.Handle("Mkdir", &ProxyHdlArgs{De: de})
	return ress.Err
}

// Stat implements [ProxyDssa].
func (pd *proxyDssa) Stat(path_ string) (*dssa.DataEntry, error) {
	ress := pd.handler.Handle("Stat", &ProxyHdlArgs{Path: path_})
	return ress.De, ress.Err
}

// SetStat implements [ProxyDssa].
func (pd *proxyDssa) SetStat(de *dssa.DataEntry, noPerm bool, noMtime bool) error {
	ress := pd.handler.Handle("SetStat", &ProxyHdlArgs{De: de, NoPerm: noPerm, NoMtime: noMtime})
	return ress.Err
}

// Checksum implements [ProxyDssa].
func (pd *proxyDssa) Checksum(algos string, path_ string) (string, error) {
	ress := pd.handler.Handle("Checksum", &ProxyHdlArgs{Algos: algos, Path: path_})
	return ress.ChSums, ress.Err
}

// GetReadCloser implements [ProxyDssa].
func (pd *proxyDssa) GetReadCloser(path_ string) (io.ReadCloser, error) {
	ress := pd.handler.Handle("GetReadCloser", &ProxyHdlArgs{Path: path_})
	return ress.RC, ress.Err
}

// GetWriteCloser implements [ProxyDssa].
func (pd *proxyDssa) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	ress := pd.handler.Handle("GetWriteCloser", &ProxyHdlArgs{Path: path_})
	return ress.WC, ress.Err
}

// Rm implements [ProxyDssa].
func (pd *proxyDssa) Rm(path_ string) error {
	ress := pd.handler.Handle("Rm", &ProxyHdlArgs{Path: path_})
	return ress.Err
}

// Symlink implements [ProxyDssa].
func (pd *proxyDssa) Symlink(old string, new_ string) error {
	ress := pd.handler.Handle("Symlink", &ProxyHdlArgs{Old: old, New: new_})
	return ress.Err
}

// Underlying implements [ProxyDssa].
func (pd *proxyDssa) Underlying() dssa.Dssa {
	return pd.underlying
}

func MakeProxyDssa(lgr *slog.Logger, underlying dssa.Dssa, handler ProxyHandler) ProxyDssa {
	if handler == nil {
		handler = &noOpHandler{lgr: lgr, underlying: underlying}
	}
	return &proxyDssa{lgr: lgr, underlying: underlying, handler: handler}
}
