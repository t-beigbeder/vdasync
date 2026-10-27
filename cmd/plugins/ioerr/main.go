package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/proxy"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"google.golang.org/grpc"
)

func RunIoErrPlugin() {
	var (
		hostFlag    = flag.String("host", "localhost", "host/address to listen, defaults to localhost")
		portFlag    = flag.Int("port", 0, "port to listen")
		nameFlag    = flag.String("name", "", "plugin name")
		typeFlag    = flag.String("type", "", "plugin type")
		resetONFlag = flag.Bool("ioeron", false, "reset errors options on new session")
		listFlag    = flag.Int("ioelist", 0, "list error rate, 100 means 1/100")
		mkdirFlag   = flag.Int("ioemkdir", 0, "list error rate, 100 means 1/100")
		statFlag    = flag.Int("ioestat", 0, "list error rate, 100 means 1/100")
		setStatFlag = flag.Int("ioesetstat", 0, "list error rate, 100 means 1/100")
		getRCFlag   = flag.Int("ioegetrc", 0, "list error rate, 100 means 1/100")
		readFlag    = flag.Int("ioeread", 0, "list error rate, 100 means 1/100")
		rCloseFlag  = flag.Int("ioerclose", 0, "list error rate, 100 means 1/100")
		getWCFlag   = flag.Int("ioegetwc", 0, "list error rate, 100 means 1/100")
		writeFlag   = flag.Int("ioewrite", 0, "list error rate, 100 means 1/100")
		wCloseFlag  = flag.Int("ioewclose", 0, "list error rate, 100 means 1/100")
		rmFlag      = flag.Int("ioerm", 0, "list error rate, 100 means 1/100")
		symLinkFlag = flag.Int("ioesymlink", 0, "list error rate, 100 means 1/100")
	)

	cf := cli.CommonFlags()
	flag.Parse()
	exe, err := os.Executable()
	if err != nil {
		common.Fatal(nil, fmt.Errorf("os.Executable: %v", err))
	}
	cmd := path.Base(exe)
	lgr, err := common.CliLogger(cmd, *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("path.Base: %s: %v", exe, err))
	}

	dss := localfiles.MakeLocalFilesDssa()
	ioeOpts := &proxy.IoErrOpts{
		ResetOnNew:         *resetONFlag,
		ListRate:           *listFlag,
		MkdirRate:          *mkdirFlag,
		StatRate:           *statFlag,
		SetStatRate:        *setStatFlag,
		GetReadCloserRate:  *getRCFlag,
		ReadRate:           *readFlag,
		RCloseRate:         *rCloseFlag,
		GetWriteCloserRate: *getWCFlag,
		WriteRate:          *writeFlag,
		WCloseRate:         *wCloseFlag,
		RmRate:             *rmFlag,
		SymlinkRate:        *symLinkFlag,
	}
	dss = proxy.MakeIoErrDss(lgr, dss, ioeOpts)

	sop, err := cli.GetServerOrPluginTls(cf)
	if err != nil {
		common.Fatal(lgr, err)
	}
	var sops []grpc.ServerOption
	if sop != nil {
		sops = []grpc.ServerOption{sop}
	}

	lgr.Info(fmt.Sprintf("%s.main starting", cmd), "version", config.GetVersion(), "name", *nameFlag, "type", *typeFlag, "host", *hostFlag, "port", *portFlag)
	done := make(chan bool)
	cb := func() {
		lgr.Debug("shutdownCb called, closing done")
		close(done)
	}
	_, _, err = remote.RunOpeDssaServer(
		lgr, context.Background(), *hostFlag, *portFlag,
		sops, dss, cb, nil)
	<-done
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("RunOpeDssaServer failed %s", err))
	}
	lgr.Info(fmt.Sprintf("%s.main done", cmd), "host", *hostFlag, "port", *portFlag)
}

func main() {
	RunIoErrPlugin()
}
