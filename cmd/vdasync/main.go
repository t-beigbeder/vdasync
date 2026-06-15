package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/plugin"
	"github.com/t-beigbeder/vdasync/internal/walker"
)

func main() {
	var (
		sourceFlag  = flag.String("source", "", "source of the command if sync")
		targetFlag  = flag.String("target", "", "target of the command")
		dryRunFlag  = flag.Bool("dryrun", false, "don't run operation, just report actions")
		rmFlag      = flag.Bool("rm", false, "remove files in sync target")
		checkFlag   = flag.Bool("check", false, "compute checksums")
		noPermFlag  = flag.Bool("noperm", false, "neither check nor set permissions")
		noMtimeFlag = flag.Bool("nomtime", false, "don't set modification time, update if source changed later")
		exclFlag    = flag.String("excl", "", "file containing regexps for paths to be excluded, defaults to none")
		inclFlag    = flag.String("incl", "", "file containing regexps for paths to be included, defaults to all")
	)
	cf := cli.CommonFlags()
	flag.Parse()
	lgr, err := common.CliLogger("vdasync", *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if *cf.VersionFlag {
		fmt.Println(config.GetVersion())
		os.Exit(0)
	}

	if *exclFlag != "" && !common.FileExists(*exclFlag) {
		common.Fatal(lgr, fmt.Errorf("exclusion file: %s does not exist", *exclFlag))
	}
	if *inclFlag != "" && !common.FileExists(*inclFlag) {
		common.Fatal(lgr, fmt.Errorf("inclusion file: %s does not exist", *inclFlag))
	}

	var rps []*plugin.RunningPlugin
	var cfg *config.CliConfig
	if *cf.ConfigFlag != "" {
		confData, err := common.LoadFile(*cf.ConfigFlag)
		if err != nil {
			common.Fatal(lgr, err)
		}
		if rps, err = cli.RunPlugins(lgr, string(confData), cf); err != nil {
			common.Fatal(lgr, err)
		}
		if len(plugin.Errors(rps)) > 0 {
			lgr.Error("some errors occured in plugins", "errs", plugin.Errors(rps))
			cli.CleanUp(lgr, rps)
			common.Fatal(lgr, errors.New("plugins error(s)"))
		}
		defer cli.CleanUp(lgr, rps)
		if cfg, err = config.Load(string(confData)); err != nil {
			common.Fatal(lgr, err)
		}
	}
	if rps != nil {
		cli.SetSignalHandler(lgr, rps)
	}

	if *sourceFlag == "" || *targetFlag == "" {
		common.Fatal(lgr, errors.New("source and target must be provided"))
	}

	sDss, sourceRoot, err := cli.GetDssAndRootFor(lgr, cf, cfg, false, *sourceFlag, rps)
	if err != nil {
		common.Fatal(lgr, err)
	}
	defer sDss.EndSession()
	tDss, targetRoot, err := cli.GetDssAndRootFor(lgr, cf, cfg, true, *targetFlag, rps)
	if err != nil {
		common.Fatal(lgr, err)
	}
	defer tDss.EndSession()

	swk, err := walker.RunSynchronizer(
		lgr, *cf.ConcurrencyFlag,
		&config.SyncOptionsType{
			Dryrun: *dryRunFlag, Rm: *rmFlag, Check: *checkFlag,
			NoPerm: *noPermFlag, NoMtime: *noMtimeFlag,
			ExclListPath: *exclFlag, InclListPath: *inclFlag,
		},
		sDss, sourceRoot,
		tDss, targetRoot,
	)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if !*cf.SilentFlag {
		syncRes := walker.SyncResult(swk)
		walker.DisplaySyncResult(syncRes, os.Stdout, true, *cf.VerboseFlag)
	}
	time.Sleep(10 * time.Millisecond)
}
