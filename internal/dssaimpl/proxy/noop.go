package proxy

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
)

type noOpHandler struct {
	lgr        *slog.Logger
	underlying dssa.Dssa
}

var _ ProxyHandler = &noOpHandler{}

func (ph *noOpHandler) Handle(fcName string, args *ProxyHdlArgs) *ProxyHdlRess {
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
		var urc io.ReadCloser
		if urc, ress.Err = ph.underlying.GetReadCloser(args.Path); ress.Err != nil {
			break
		}
		ress.RC = makePReader(ph, urc)
	case "Read":
		ress.NRdWr, ress.Err = args.RC.Read(args.Buf)
	case "RClose":
		ress.Err = args.RC.Close()
	case "GetWriteCloser":
		var uwc io.WriteCloser
		if uwc, ress.Err = ph.underlying.GetWriteCloser(args.Path); ress.Err != nil {
			break
		}
		ress.WC = makePWriter(ph, uwc)
	case "Write":
		ress.NRdWr, ress.Err = args.WC.Write(args.Buf)
	case "WClose":
		ress.Err = args.WC.Close()
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
