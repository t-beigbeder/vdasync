package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/encrypted"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"google.golang.org/grpc"
)

func RunEncryptPlugin() {
	var (
		hostFlag       = flag.String("host", "localhost", "host/address to listen, defaults to localhost")
		portFlag       = flag.Int("port", 0, "port to listen")
		nameFlag       = flag.String("name", "", "plugin name")
		typeFlag       = flag.String("type", "", "plugin type")
		ageIdfFlag     = flag.String("ageidf", "", "age identities (secrets) file name")
		ageRecfFlag    = flag.String("agerecf", "", "age recipients (public keys) file name")
		underlyingFlag = flag.String("underlying", "", "DSS URL for encrypted files storage")
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
	if *ageRecfFlag == "" {
		common.Fatal(lgr, errors.New("agerecf empty"))
	}
	recipients, err := common.FileLines(*ageRecfFlag)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("recf: %s: %v", *ageRecfFlag, err))
	}
	if *underlyingFlag == "" {
		common.Fatal(lgr, errors.New("underlying empty"))
	}

	pName, host, port, rootPath, err := cli.ParseUrl(*underlyingFlag)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("underlying DSS parsing: %s: %v", *underlyingFlag, err))
	}
	if pName != "" {
		common.Fatal(lgr, fmt.Errorf("encrypt underlying does not support plugins at the moment (%s)", *underlyingFlag))
	}
	var underlying dssa.Dssa
	if host == "" && port == 0 {
		underlying = localfiles.MakeLocalFilesDssa()
	} else {
		underlying, err = cli.GetGrpcClient(lgr, cf, host, port)
		if err != nil {
			common.Fatal(lgr, fmt.Errorf("cli.GetGrpcClient: %s: %v", *underlyingFlag, err))
		}
		defer underlying.EndSession()
	}

	dss, err := encrypted.MakeEncryptedDssa(lgr, underlying, rootPath, identities, recipients)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("encrypted.MakeEncryptedDssa: %s: %v", exe, err))
	}

	sop, err := cli.GetServerOrPluginTls(cf)
	if err != nil {
		common.Fatal(lgr, err)
	}
	var sops []grpc.ServerOption
	if sop != nil {
		sops = []grpc.ServerOption{sop}
	}

	lgr.Info(fmt.Sprintf("%s.main starting", cmd), "name", *nameFlag, "type", *typeFlag, "host", *hostFlag, "port", *portFlag)
	done := make(chan bool)
	cb := func() {
		lgr.Debug("shutdownCb called, closing done")
		close(done)
	}
	_, _, err = remote.RunOpeDssaServer(
		lgr, context.Background(), *hostFlag, *portFlag,
		sops, dss, cb)
	<-done
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("RunOpeDssaServer failed %s", err))
	}
	lgr.Info(fmt.Sprintf("%s.main done", cmd), "host", *hostFlag, "port", *portFlag)
}

func main() {
	RunEncryptPlugin()
}
