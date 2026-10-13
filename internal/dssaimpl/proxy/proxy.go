package proxy

import (
	"fmt"
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
}

type ProxyHdlRess struct {
	Err    error
	Des []*dssa.DataEntry
	De     *dssa.DataEntry
	RC     io.ReadCloser
	WC     io.WriteCloser
	ChSums string
}

type ProxyHandler interface {
	Handle(fcName string, args *ProxyHdlArgs) *ProxyHdlRess
}

type proxyDssa struct {
	lgr *slog.Logger
	underlying dssa.Dssa
	handler ProxyHandler
}

type noOpHandler struct {
	lgr *slog.Logger
	underlying dssa.Dssa
}

func (ph *noOpHandler)Handle(fcName string, args *ProxyHdlArgs) *ProxyHdlRess {
	ress := &ProxyHdlRess{}
	switch fcName {
	case "NewSession":
		ress.Err = ph.underlying.NewSession()
	case "EndSession":
		ress.Err = ph.underlying.EndSession()
	case "List":
		ress.Des, ress.Err = ph.underlying.List(args.Path)
	case "Mkdir":
		ress.Err = ph.underlying.Mkdir(args.De)
	case "Stat":
		ress.De, ress.Err = ph.underlying.Stat(args.Path)
	case "SetStat":
		ress.Err = ph.underlying.SetStat(args.De, args.NoPerm, args.NoMtime)
	case "Checksum":
		ress.ChSums, ress.Err = ph.underlying.Checksum(args.Algos, args.Path)
	case "GetReadCloser":
		ress.RC, ress.Err = ph.underlying.GetReadCloser(args.Path)
	case "GetWriteCloser":
		ress.WC, ress.Err = ph.underlying.GetWriteCloser(args.Path)
	case "Rm":
		ress.Err = ph.underlying.Rm(args.Path)
	case "Symlink":
		ress.Err = ph.underlying.Symlink(args.Old, args.New)
	default:
		ress.Err = fmt.Errorf("invalid function name %s to handle", fcName)
	}
	dep := ""
	if args.De != nil {
		dep = args.De.Path
	}
	rdep := ""
	if ress.De != nil {
		rdep = ress.De.Path
	}
	ph.lgr.Debug(fcName, "path", args.Path, "de", dep, "err", ress.Err, "rde", rdep, "des", len(ress.Des))
	return ress
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
