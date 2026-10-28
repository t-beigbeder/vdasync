package icapsftp

import (
	"fmt"
	"io"
		"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/proxy"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/sftpc"
	"github.com/t-beigbeder/vdasync/internal/sftputil"
)

type icapSftpHandler struct {
	lgr        *slog.Logger
	underlying dssa.Dssa
}

var _ proxy.ProxyHandler = &icapSftpHandler{}

func (ph *icapSftpHandler) Handle(fcName string, args *proxy.ProxyHdlArgs) *proxy.ProxyHdlRess {
	ress := &proxy.ProxyHdlRess{}
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
	ph.lgr.Debug(fmt.Sprintf("icapSftp: %s", fcName), "path", args.Path, "de", dep, "err", ress.Err, "rde", rdep, "des", len(ress.Des))
	return ress
}

func MakeIcapSftpDssa(
	lgr *slog.Logger,
	user, address, identity, root string,
	concurrency int,
	factory sftpc.SftpClientFactory, knownHostsFile string,
) (dssa.Dssa, error) {
	underlying, err := sftpc.MakeSftpClientDssa(
		user, address, identity, root,
		concurrency, sftputil.GetSftpClient, knownHostsFile)
	if err != nil {
		return nil, err
	}
	pd := proxy.MakeProxyDssa(
		lgr, underlying,
		&icapSftpHandler{lgr: lgr, underlying: underlying},
	)
	return pd, nil
}
