package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/plugin"
	"github.com/t-beigbeder/vdasync/internal/walker"
)

func RunSyncCli(df *DssaFactory) {
	var (
		sourceFlag   = flag.String("source", "", "source of the command")
		targetFlag   = flag.String("target", "", "target of the command")
		dryRunFlag   = flag.Bool("dryrun", false, "don't run operation, just report actions")
		rmFlag       = flag.Bool("rm", false, "remove files in sync target")
		forceFlag    = flag.Bool("force", false, "force removing read-only files in sync target")
		noPermFlag   = flag.Bool("noperm", false, "neither check nor set permissions")
		noMtimeFlag  = flag.Bool("nomtime", false, "don't set modification time, update if source changed later")
		noMtLinkFlag = flag.Bool("nomtlink", false, "same as nomtime but only applies to symlinks")
		cProfFlag    = flag.String("cprof", "", "cpu.prof file")
	)
	cf := CommonFlags()
	svsf := ServicesFlags()
	flag.Parse()
	lgr, err := common.CliLogger("vdasync", *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if *cProfFlag != "" {
		cpuPf, err := os.Create(*cProfFlag)
		if err != nil {
			common.Fatal(lgr, fmt.Errorf("cprof %s: %v", *cProfFlag, err))
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

	if *svsf.ExclFlag != "" && !common.FileExists(*svsf.ExclFlag) {
		common.Fatal(lgr, fmt.Errorf("exclusion file: %s does not exist", *svsf.ExclFlag))
	}
	if *svsf.InclFlag != "" && !common.FileExists(*svsf.InclFlag) {
		common.Fatal(lgr, fmt.Errorf("inclusion file: %s does not exist", *svsf.InclFlag))
	}

	var rps []*plugin.RunningPlugin
	cfg, err := LoadConfig(cf)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if cfg != nil {
		if rps, err = RunPlugins(lgr, cfg, cf); err != nil {
			common.Fatal(lgr, err)
		}
		if len(plugin.Errors(rps)) > 0 {
			lgr.Error("some errors occured in plugins", "errs", plugin.Errors(rps))
			CleanUp(lgr, rps)
			common.Fatal(lgr, errors.New("plugins error(s)"))
		}
		defer CleanUp(lgr, rps)
	}
	if rps != nil {
		SetSignalHandler(lgr, rps)
	}

	if *sourceFlag == "" || *targetFlag == "" {
		common.Fatal(lgr, errors.New("source and target must be provided"))
	}

	sDss, sourceRoot, err := GetDssAndRootFor(lgr, cf, cfg, false, *sourceFlag, rps, df)
	if err != nil {
		common.Fatal(lgr, err)
	}
	defer sDss.EndSession()
	tDss, targetRoot, err := GetDssAndRootFor(lgr, cf, cfg, true, *targetFlag, rps, df)
	if err != nil {
		common.Fatal(lgr, err)
	}
	defer tDss.EndSession()

	swk, err := walker.RunSynchronizer(
		lgr, *cf.ConcurrencyFlag,
		&config.SyncOptionsType{
			Dryrun: *dryRunFlag, Rm: *rmFlag, Force: *forceFlag,
			Check: *svsf.CheckFlag, CsAlgos: *svsf.CsalFlag,
			NoPerm: *noPermFlag, NoMtime: *noMtimeFlag, NoMtLink: *noMtLinkFlag,
			ExclListPath: *svsf.ExclFlag, InclListPath: *svsf.InclFlag,
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
	} else if *cf.VerboseFlag {
		walker.DisplaySyncResult(syncRes, outFile, true, false)
	}
	time.Sleep(10 * time.Millisecond)
}
