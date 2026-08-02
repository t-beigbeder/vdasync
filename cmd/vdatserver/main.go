package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/encrypted"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"google.golang.org/grpc"
)

func RunTrustedServer() {
	var (
		hostFlag   = flag.String("host", "localhost", "host/address to listen, defaults to localhost")
		portFlag   = flag.Int("port", 0, "port to listen")
		ageIdfFlag = flag.String("ageidf", "", "age identities (secrets) file name")
		rootFlag   = flag.String("root", "", "root path for encrypted files storage")
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

	if *ageIdfFlag == "" {
		common.Fatal(lgr, errors.New("ageidf empty"))
	}
	identities, err := common.FileLines(*ageIdfFlag)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("idf: %s: %v", *ageIdfFlag, err))
	}
	if *rootFlag == "" {
		common.Fatal(lgr, errors.New("root empty"))
	}

	dss, err := encrypted.MakeProxyDssa(lgr, *rootFlag, identities)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("encrypted.MakeProxyDssa: %s: %v", exe, err))
	}

	sop, err := cli.GetServerOrPluginTls(cf)
	if err != nil {
		common.Fatal(lgr, err)
	}
	var sops []grpc.ServerOption
	if sop != nil {
		sops = []grpc.ServerOption{sop}
	}

	lgr.Info(fmt.Sprintf("%s.main starting", cmd), "version", config.GetVersion(), "host", *hostFlag, "port", *portFlag)
	done := make(chan bool)
	cb := func() {
		lgr.Debug("shutdownCb called, closing done")
		close(done)
	}
	_, _, err = remote.RunOpeDssaServer(
		lgr, context.Background(), *hostFlag, *portFlag,
		sops, dss, cb, dss.GetValueSetCb())
	<-done
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("RunOpeDssaServer failed %s", err))
	}
	dss.StopSessionMonitor()
	lgr.Info(fmt.Sprintf("%s.main done", cmd), "host", *hostFlag, "port", *portFlag)
}

func main() {
	RunTrustedServer()
}
