package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/plugin"
	"github.com/t-beigbeder/vdasync/internal/walker"
)

func main() {
	var (
		sourceFlag   = flag.String("source", "", "source of the command")
		targetFlag   = flag.String("target", "", "target of the command")
		dryRunFlag   = flag.Bool("dryrun", false, "don't run operation, just report actions")
		rmFlag       = flag.Bool("rm", false, "remove files in sync target")
		checkFlag    = flag.Bool("check", false, "compute checksums")
		noPermFlag   = flag.Bool("noperm", false, "neither check nor set permissions")
		noMtimeFlag  = flag.Bool("nomtime", false, "don't set modification time, update if source changed later")
		noMtLinkFlag = flag.Bool("nomtlink", false, "same as nomtime but only applies to symlinks")
		exclFlag     = flag.String("excl", "", "file containing regexps for paths to be excluded, defaults to none")
		inclFlag     = flag.String("incl", "", "file containing regexps for paths to be included, defaults to all")
		cProfFlag    = flag.String("cprof", "", "cpu.prof file")
	)
	cf := cli.CommonFlags()
	flag.Parse()
	lgr, err := common.CliLogger("vdasync", *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if *cProfFlag != "" {
		cpuPf, err := os.Create(*cProfFlag)
		if err != nil {
			common.Fatal(lgr, fmt.Errorf("cprof %s: %v", *cpuPf, err))
		}
		defer cpuPf.Close()
		if err := pprof.StartCPUProfile(cpuPf); err != nil {
			common.Fatal(lgr, fmt.Errorf("StartCPUProfile: %v", err))
		}
		defer pprof.StopCPUProfile()
	}
	if *cf.VersionFlag {
		fmt.Println(config.GetVersion())
		os.Exit(0)
	}
	outFile, err := common.StdWriter(*cf.OutFlag)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("output file %s: %v", *cf.OutFlag, err))
	}
	defer outFile.Close()

	if *exclFlag != "" && !common.FileExists(*exclFlag) {
		common.Fatal(lgr, fmt.Errorf("exclusion file: %s does not exist", *exclFlag))
	}
	if *inclFlag != "" && !common.FileExists(*inclFlag) {
		common.Fatal(lgr, fmt.Errorf("inclusion file: %s does not exist", *inclFlag))
	}

	var rps []*plugin.RunningPlugin
	cfg, err := cli.LoadConfig(cf)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if cfg != nil {
		if rps, err = cli.RunPlugins(lgr, cfg, cf); err != nil {
			common.Fatal(lgr, err)
		}
		if len(plugin.Errors(rps)) > 0 {
			lgr.Error("some errors occured in plugins", "errs", plugin.Errors(rps))
			cli.CleanUp(lgr, rps)
			common.Fatal(lgr, errors.New("plugins error(s)"))
		}
		defer cli.CleanUp(lgr, rps)
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
			NoPerm: *noPermFlag, NoMtime: *noMtimeFlag, NoMtLink: *noMtLinkFlag,
			ExclListPath: *exclFlag, InclListPath: *inclFlag,
		},
		sDss, sourceRoot,
		tDss, targetRoot,
	)
	if err != nil {
		common.Fatal(lgr, err)
	}
	syncRes := walker.SyncResult(swk)
	if !*cf.SilentFlag {
		walker.DisplaySyncResult(syncRes, outFile, true, *cf.VerboseFlag)
	} else if !*cf.VerboseFlag {
		walker.DisplaySyncResult(syncRes, outFile, true, false)
	}
	time.Sleep(10 * time.Millisecond)
}
