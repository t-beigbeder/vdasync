package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/walker"
)

func setSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Fprintf(os.Stderr, "signal %s received, preparing to exit\n", sig)
		}
	}()
}

func main() {
	var (
		sourceFlag  = flag.String("source", "", "source of the command if sync")
		targetFlag  = flag.String("target", "", "target of the command")
		dryRunFlag  = flag.Bool("dryrun", false, "don't run operation, just report actions")
		rmFlag      = flag.Bool("rm", false, "remove files in sync target")
		checkFlag   = flag.Bool("check", false, "compute checksums")
		noPermFlag  = flag.Bool("noperm", false, "neither check nor set permissions")
		noMtimeFlag = flag.Bool("nomtime", false, "don't set modification time, update if source changed later")
	)
	setSignalHandler()
	time.Sleep(10*time.Second)
	cf := cli.CommonFlags()
	flag.Parse()
	lgr, err := common.CliLogger("vdasync", *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	lgr = lgr.With("app", "vdasync")
	if *sourceFlag == "" || *targetFlag == "" {
		common.Fatal(lgr, errors.New("source and target must be provided"))
	}
	sourceRoot, err := cli.NormalizeRoot(*sourceFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	targetRoot, err := cli.NormalizeRoot(*targetFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	dss := localfiles.MakeLocalFilesDssa()
	swk, err := walker.RunSynchronizer(
		lgr, *cf.ConcurrencyFlag,
		&config.SyncOptionsType{
			Dryrun: *dryRunFlag, Rm: *rmFlag, Check: *checkFlag,
			NoPerm: *noPermFlag, NoMtime: *noMtimeFlag,
		},
		dss, sourceRoot,
		dss, targetRoot,
	)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if !*cf.SilentFlag {
		syncRes := walker.SyncResult(swk)
		walker.DisplaySyncResult(syncRes, os.Stdout, true, *cf.VerboseFlag)
	}
}
