package proxy

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
)

type RetryErrOpts struct {
	max int
}

type retryErrHandler struct {
	lgr        *slog.Logger
	underlying dssa.Dssa
	opts       *RetryErrOpts
}

var _ ProxyHandler = &retryErrHandler{}

func (ph *retryErrHandler) Handle(fcName string, args *ProxyHdlArgs) *ProxyHdlRess {
	ress := &ProxyHdlRess{}
	doUntil := func(f func()) {
		for count := range ph.opts.max + 1 {
			f()
			if ress.Err == nil || ress.Err == io.EOF {
				if count != 0 {
					ph.lgr.Debug("retry succeeded", "fcName", fcName, "count", count, "max", ph.opts.max+1)
				}
				return
			}
			ph.lgr.Debug("retry pending", "fcName", fcName, "err", ress.Err.Error(), "count", count, "max", ph.opts.max+1)
		}
		ph.lgr.Debug("retry failed", "fcName", fcName, "err", ress.Err.Error(), "max", ph.opts.max+1)
	}
	switch fcName {
	case "NewSession":
		ress.Err = ph.underlying.NewSession()
	case "EndSession":
		ress.Err = ph.underlying.EndSession()
	case "List":
		doUntil(func() {
			ress.Des, ress.Err = ph.underlying.List(args.Path)
		})
	case "Mkdir":
		doUntil(func() {
			ress.Err = ph.underlying.Mkdir(args.De)
		})
	case "Stat":
		doUntil(func() {
			ress.De, ress.Err = ph.underlying.Stat(args.Path)
		})
	case "SetStat":
		doUntil(func() {
			ress.Err = ph.underlying.SetStat(args.De, args.NoPerm, args.NoMtime)
		})
	case "Checksum":
		ress.ChSums, ress.Err = ph.underlying.Checksum(args.Algos, args.Path)
	case "GetReadCloser":
		doUntil(func() {
			var urc io.ReadCloser
			urc, ress.Err = ph.underlying.GetReadCloser(args.Path)
			if ress.Err == nil {
				ress.RC = makePReader(ph, urc)
			}
		})
	case "Read":
		doUntil(func() {
			ress.NRdWr, ress.Err = args.RC.Read(args.Buf)
		})
	case "RClose":
		doUntil(func() {
			ress.Err = args.RC.Close()
		})
	case "GetWriteCloser":
		doUntil(func() {
			var uwc io.WriteCloser
			uwc, ress.Err = ph.underlying.GetWriteCloser(args.Path)
			if ress.Err == nil {
				ress.WC = makePWriter(ph, uwc)
			}
		})
	case "Write":
		doUntil(func() {
			ress.NRdWr, ress.Err = args.WC.Write(args.Buf)
		})
	case "WClose":
		doUntil(func() {
			ress.Err = args.WC.Close()
		})
	case "Rm":
		doUntil(func() {
			ress.Err = ph.underlying.Rm(args.Path)
		})
	case "Symlink":
		doUntil(func() {
			ress.Err = ph.underlying.Symlink(args.Old, args.New)
		})
	default:
		ress.Err = fmt.Errorf("invalid functretryn name %s to handle", fcName)
	}
	dep := ""
	if args.De != nil {
		dep = args.De.Path
	}
	rdep := ""
	if ress.De != nil {
		rdep = ress.De.Path
	}
	ph.lgr.Debug(fmt.Sprintf("retryErr: %s", fcName), "path", args.Path, "de", dep, "err", ress.Err, "rde", rdep, "des", len(ress.Des))
	return ress
}

func MakeRetryErrHandler(lgr *slog.Logger, underlying dssa.Dssa, opts *RetryErrOpts) ProxyHandler {
	ph := &retryErrHandler{
		lgr:        lgr,
		underlying: underlying,
		opts:       opts,
	}
	return ph
}

type retryeDssa struct {
	*proxyDssa
}

func MakeRetryErrDss(lgr *slog.Logger, underlying dssa.Dssa, opts *RetryErrOpts) dssa.Dssa {
	if opts == nil {
		opts = &RetryErrOpts{}
	}
	handler := MakeRetryErrHandler(lgr, underlying, opts)
	dss := &retryeDssa{
		proxyDssa: &proxyDssa{lgr: lgr, underlying: underlying, handler: handler},
	}
	return dss
}
