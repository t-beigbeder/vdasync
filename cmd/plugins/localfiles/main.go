package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/remote"
)

func main() {
	var (
		hostFlag = flag.String("host", "localhost", "host/address to listen, defaults to localhost")
		portFlag = flag.Int("port", 0, "port to listen")
		nameFlag = flag.String("name", "", "plugin name")
		typeFlag = flag.String("type", "", "plugin type")
	)
	cf := cli.CommonFlags()
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
	_, _ = *nameFlag, *typeFlag

	lgr.Info(fmt.Sprintf("%s.main starting", cmd), "host", *hostFlag, "port", *portFlag)
	done := make(chan bool)
	cb := func() {
		lgr.Debug("shutdownCb called, closing done")
		close(done)
	}
	_, _, err = remote.RunOpeDssaServer(context.Background(), *hostFlag, *portFlag, nil, localfiles.MakeLocalFilesDssa(), cb)
	<-done
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("RunOpeDssaServer failed %s", err))
	}
	lgr.Info(fmt.Sprintf("%s.main done", cmd), "host", *hostFlag, "port", *portFlag)
}
