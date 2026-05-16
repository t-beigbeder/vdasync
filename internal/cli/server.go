package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"google.golang.org/grpc"
)

func RunServerOrPlugin(isPlugin bool) {
	var (
		hostFlag = flag.String("host", "localhost", "host/address to listen, defaults to localhost")
		portFlag = flag.Int("port", 0, "port to listen")
		nameFlag *string
		typeFlag *string
	)
	if isPlugin {
		nameFlag = flag.String("name", "", "plugin name")
		typeFlag = flag.String("type", "", "plugin type")
	}
	cf := CommonFlags()
	flag.Parse()
	exe, err := os.Executable()
	if err != nil {
		common.Fatal(nil, err)
	}
	cmd := path.Base(exe)
	lgr, err := common.CliLogger(cmd, *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	sop, err := GetServerOrPluginTls(cf)
	if err != nil {
		common.Fatal(lgr, err)
	}
	var sops []grpc.ServerOption
	if sop != nil {
		sops = []grpc.ServerOption{sop}
	}

	if isPlugin {
		lgr.Info(fmt.Sprintf("%s.main starting", cmd), "name", *nameFlag, "type", *typeFlag, "host", *hostFlag, "port", *portFlag)
	} else {
		lgr.Info(fmt.Sprintf("%s.main starting", cmd), "host", *hostFlag, "port", *portFlag)
	}
	done := make(chan bool)
	cb := func() {
		lgr.Debug("shutdownCb called, closing done")
		close(done)
	}
	_, _, err = remote.RunOpeDssaServer(
		lgr, context.Background(), *hostFlag, *portFlag,
		sops, localfiles.MakeLocalFilesDssa(), cb)
	<-done
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("RunOpeDssaServer failed %s", err))
	}
	lgr.Info(fmt.Sprintf("%s.main done", cmd), "host", *hostFlag, "port", *portFlag)

}
