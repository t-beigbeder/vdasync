package proxy

import (
	"fmt"
	"io"
	"log/slog"
	mrand "math/rand"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
)

type IoErrOpts struct {
	ListRate           int
	MkdirRate          int
	StatRate           int
	SetStatRate        int
	GetReadCloserRate  int
	GetWriteCloserRate int
	RmRate             int
	SymlinkRate        int
}

type ioErrHandler struct {
	lgr        *slog.Logger
	underlying dssa.Dssa
	opts       *IoErrOpts
}

var _ ProxyHandler = &ioErrHandler{}

func (ph *ioErrHandler) Handle(fcName string, args *ProxyHdlArgs) *ProxyHdlRess {
	ress := &ProxyHdlRess{}
	doBeforeWithError := func(rate int, f func()) {
		f()
		if rate == 0 {
			return
		}
		if ress.Err != nil {
			ph.lgr.Debug("ioerr before useless", "err", ress.Err)
			return
		}
		if mrand.Intn(rate) == 0 {
			v := time.Now().UnixNano()
			ress.Err = fmt.Errorf("rand %d 1/%d", v, ph.opts.ListRate)
			return
		}
	}
	doAfterWithError := func(rate int, f func()) {
		if rate == 0 {
			f()
			return
		}
		if mrand.Intn(rate) == 0 {
			v := time.Now().UnixNano()
			ress.Err = fmt.Errorf("rand %d 1/%d", v, ph.opts.ListRate)
			return
		}
		ph.lgr.Debug("ioerr after", "err", ress.Err)
		f()
	}
	switch fcName {
	case "NewSession":
		ress.Err = ph.underlying.NewSession()
	case "EndSession":
		ress.Err = ph.underlying.EndSession()
	case "List":
		doBeforeWithError(ph.opts.ListRate, func() {
			ress.Des, ress.Err = ph.underlying.List(args.Path)
		})
	case "Mkdir":
		doAfterWithError(ph.opts.MkdirRate, func() {
			ress.Err = ph.underlying.Mkdir(args.De)
		})
	case "Stat":
		doBeforeWithError(ph.opts.StatRate, func() {
			ress.De, ress.Err = ph.underlying.Stat(args.Path)
		})
	case "SetStat":
		doAfterWithError(ph.opts.SetStatRate, func() {
			ress.Err = ph.underlying.SetStat(args.De, args.NoPerm, args.NoMtime)
		})
	case "Checksum":
		ress.ChSums, ress.Err = ph.underlying.Checksum(args.Algos, args.Path)
	case "GetReadCloser":
		doAfterWithError(ph.opts.GetReadCloserRate, func() {
			var urc io.ReadCloser
			urc, ress.Err = ph.underlying.GetReadCloser(args.Path)
			if ress.Err == nil {
				ress.RC = makePReader(ph, urc)
			}
		})
	case "Read":
		ress.NRdWr, ress.Err = args.RC.Read(args.Buf)
	case "RClose":
		ress.Err = args.RC.Close()
	case "GetWriteCloser":
		doAfterWithError(ph.opts.GetWriteCloserRate, func() {
			var uwc io.WriteCloser
			uwc, ress.Err = ph.underlying.GetWriteCloser(args.Path)
			if ress.Err == nil {
				ress.WC = makePWriter(ph, uwc)
			}
		})
	case "Write":
		ress.NRdWr, ress.Err = args.WC.Write(args.Buf)
	case "WClose":
		ress.Err = args.WC.Close()
	case "Rm":
		doAfterWithError(ph.opts.RmRate, func() {
			ress.Err = ph.underlying.Rm(args.Path)
		})
	case "Symlink":
		doAfterWithError(ph.opts.SymlinkRate, func() {
			ress.Err = ph.underlying.Symlink(args.Old, args.New)
		})
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
	ph.lgr.Debug(fmt.Sprintf("ioErr: %s", fcName), "path", args.Path, "de", dep, "err", ress.Err, "rde", rdep, "des", len(ress.Des))
	return ress
}

func MakeIoErrHandler(lgr *slog.Logger, underlying dssa.Dssa, opts *IoErrOpts) ProxyHandler {
	ph := &ioErrHandler{
		lgr:        lgr,
		underlying: underlying,
		opts:       opts,
	}
	return ph
}

type ioeDssa struct {
	*proxyDssa
}

func MakeIoErrDss(lgr *slog.Logger, underlying dssa.Dssa, opts *IoErrOpts) dssa.Dssa {
	if opts == nil {
		opts = &IoErrOpts{}
	}
	handler := MakeIoErrHandler(lgr, underlying, opts)
	dss := &ioeDssa{
		proxyDssa: &proxyDssa{lgr: lgr, underlying: underlying, handler: handler},
	}
	return dss
}
