package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

type rateLimHandler struct {
	lgr        *slog.Logger
	underlying dssa.Dssa
	Ebl        *common.EvBdwLter
}

// GetEbl implements [RateLimProxyHandler].
func (ph *rateLimHandler) GetEbl() *common.EvBdwLter {
	return ph.Ebl
}

type RateLimProxyHandler interface {
	ProxyHandler
	GetEbl() *common.EvBdwLter
}

type RateLimitedDss interface {
	dssa.Dssa
	GetEbl() *common.EvBdwLter
}

var _ RateLimProxyHandler = &rateLimHandler{}

func (ph *rateLimHandler) Handle(fcName string, args *ProxyHdlArgs) *ProxyHdlRess {
	ress := &ProxyHdlRess{}
	switch fcName {
	case "NewSession":
		ress.Err = ph.underlying.NewSession()
	case "EndSession":
		ress.Err = ph.underlying.EndSession()
	case "List":
		ph.Ebl.PondeRtlThis(50, 0)
		ress.Des, ress.Err = ph.underlying.List(args.Path)
	case "Mkdir":
		ph.Ebl.PondeRtlThis(100, 0)
		ress.Err = ph.underlying.Mkdir(args.De)
	case "Stat":
		ph.Ebl.PondeRtlThis(10, 0)
		ress.De, ress.Err = ph.underlying.Stat(args.Path)
	case "SetStat":
		ph.Ebl.PondeRtlThis(50, 0)
		ress.Err = ph.underlying.SetStat(args.De, args.NoPerm, args.NoMtime)
	case "Checksum":
		ress.ChSums, ress.Err = ph.underlying.Checksum(args.Algos, args.Path)
	case "GetReadCloser":
		ph.Ebl.PondeRtlThis(20, 0)
		var urc io.ReadCloser
		if urc, ress.Err = ph.underlying.GetReadCloser(args.Path); ress.Err != nil {
			break
		}
		ress.RC = makePReader(ph, urc)
	case "Read":
		ress.NRdWr, ress.Err = args.RC.Read(args.Buf)
		if err := ph.Ebl.PondeRtlThis(1, ress.NRdWr); err != nil {
			ph.lgr.Error("rate limiter", "err", err)
			if ress.Err == nil {
				ress.Err = err
			}
		}
	case "RClose":
		ph.Ebl.PondeRtlThis(10, 0)
		ress.Err = args.RC.Close()
	case "GetWriteCloser":
		ph.Ebl.PondeRtlThis(40, 0)
		var uwc io.WriteCloser
		if uwc, ress.Err = ph.underlying.GetWriteCloser(args.Path); ress.Err != nil {
			break
		}
		ress.WC = makePWriter(ph, uwc)
	case "Write":
		ress.NRdWr, ress.Err = args.WC.Write(args.Buf)
		if err := ph.Ebl.PondeRtlThis(2, ress.NRdWr); err != nil {
			ph.lgr.Error("rate limiter", "err", err)
			if ress.Err == nil {
				ress.Err = err
			}
		}
	case "WClose":
		ph.Ebl.PondeRtlThis(20, 0)
		ress.Err = args.WC.Close()
	case "Rm":
		ph.Ebl.PondeRtlThis(50, 0)
		ress.Err = ph.underlying.Rm(args.Path)
	case "Symlink":
		ph.Ebl.PondeRtlThis(50, 0)
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
	ph.lgr.Debug(fmt.Sprintf("rateLim: %s", fcName), "path", args.Path, "de", dep, "err", ress.Err, "rde", rdep, "des", len(ress.Des))
	return ress
}

func MakeRateLimHandler(lgr *slog.Logger, underlying dssa.Dssa, evFreq float64, bdw float64) RateLimProxyHandler {
	rlh := &rateLimHandler{
		lgr:        lgr,
		underlying: underlying,
	}
	rlh.Ebl = common.NewEvBdwLter(evFreq, bdw,
		func(kw string, d time.Duration) {
			switch kw {
			case "":
				rlh.lgr.Debug("event without delay")
			case "e":
				rlh.lgr.Debug("event rate limited", "delay", d)
			case "b":
				rlh.lgr.Debug("bandwidth limited", "delay", d)
			}
		},
	)
	return rlh
}

type rlDssa struct {
	*proxyDssa
	rlh RateLimProxyHandler
}

// GetEbl implements [RateLimitedDss].
func (r *rlDssa) GetEbl() *common.EvBdwLter {
	return r.rlh.GetEbl()
}

func MakeRateLimitedDss(lgr *slog.Logger, underlying dssa.Dssa, evr float64, bdw float64) RateLimitedDss {
	handler := MakeRateLimHandler(lgr, underlying, evr, bdw)
	dss := &rlDssa{
		proxyDssa: &proxyDssa{lgr: lgr, underlying: underlying, handler: handler},
		rlh:       handler,
	}
	dss.GetEbl().Record()
	return dss
}
