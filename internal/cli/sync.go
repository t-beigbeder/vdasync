package cli

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/opelogimpl"
	"github.com/t-beigbeder/vdasync/internal/plugin"
	"github.com/t-beigbeder/vdasync/internal/walker"
	"github.com/t-beigbeder/vdasync/opelog"
)

func RunSyncCli(df *DssaFactory) {
	var (
		oplCmdFlag     = flag.String("oplcmd", "", "operation logs command: load,create,update")
		oplNoInvCFlag  = flag.Bool("oplnoinvc", false, "operation logs won't check inventory")
		oplRtoFlag     = flag.Int64("oplrto", 0, "operation logs reset timeout in milliseconds")
		oplRestartFlag = flag.Bool("oplrestart", false, "operation logs self restart enabled")
		oplClErrFlag   = flag.Bool("oplclerr", false, "operation logs clear errors on (re)start")
		oplKeepCFlag   = flag.Bool("oplkeepc", false, "operation logs keep cache on (re)start")
		oplErrRtlFlag  = flag.String("oplertl", "1.67e-2", "operation logs errors rate limit")
		sourceFlag     = flag.String("source", "", "source of the command")
		targetFlag     = flag.String("target", "", "target of the command")
		dryRunFlag     = flag.Bool("dryrun", false, "don't run operation, just report actions")
		rmFlag         = flag.Bool("rm", false, "remove files in sync target")
		forceFlag      = flag.Bool("force", false, "force removing read-only files in sync target")
		iirregFlag     = flag.Bool("iirreg", false, "ignore irregular files (sockets, pipes, devices...)")
		noPermFlag     = flag.Bool("noperm", false, "neither check nor set permissions")
		noMtimeFlag    = flag.Bool("nomtime", false, "don't set modification time, update if source changed later")
		noMtLinkFlag   = flag.Bool("nomtlink", false, "same as nomtime but only applies to symlinks")
		cProfFlag      = flag.String("cprof", "", "cpu.prof file")
		legacyFlag     = flag.Bool("legacy", false, "won't use operation logs algorithm but legacy one, no more maintained")
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
			CleanupPlugins(lgr, rps)
			common.Fatal(lgr, errors.New("plugins error(s)"))
		}
		defer CleanupPlugins(lgr, rps)
	}
	sh := MakeSignalHandler(lgr)
	sh.RegisterCleanupFunction(func(lgr *slog.Logger) error {
		CleanupPlugins(lgr, rps)
		return nil
	})

	if *sourceFlag == "" || *targetFlag == "" {
		common.Fatal(lgr, errors.New("source and target must be provided"))
	}

	sDss, sourceRoot, err := GetDssAndRootFor(lgr, cf, cfg, false, *sourceFlag, rps, df, *svsf.SrcRtlFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	defer sDss.EndSession()
	tDss, targetRoot, err := GetDssAndRootFor(lgr, cf, cfg, true, *targetFlag, rps, df, *svsf.TgtRtlFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	defer tDss.EndSession()

	swo := config.SyncOptionsType{
		Dryrun:       *dryRunFlag,
		Rm:           *rmFlag,
		Force:        *forceFlag,
		IgnoreIrreg:  *iirregFlag,
		Check:        *svsf.CheckFlag,
		CsAlgos:      *svsf.CsalFlag,
		NoPerm:       *noPermFlag,
		NoMtime:      *noMtimeFlag,
		NoMtLink:     *noMtLinkFlag,
		ExclListPath: *svsf.ExclFlag,
		InclListPath: *svsf.InclFlag,
	}

	if *legacyFlag {
		swk, err := walker.RunSynchronizer(
			lgr, *cf.ConcurrencyFlag,
			&swo,
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
		return
	}
	if *oplCmdFlag == "" {
		// to replace legacy algorithm
		*svsf.OplPathFlag = ""
		*svsf.OplSyncFlag = ""
		*oplCmdFlag = "update"
		*oplRtoFlag = 250
		*oplNoInvCFlag = true
		*oplErrRtlFlag = ""

	} else if *svsf.OplPathFlag == "" {
		common.Fatal(lgr, errors.New("oplpath flag is missing"))
	}

	syncPer, syncBck, err := ParsePeriodBackups(*svsf.OplSyncFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	expPer, expBck, err := ParsePeriodBackups(*svsf.OplExpFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	oplm, err := opelogimpl.MakeM2fManager(*svsf.OplPathFlag, syncBck)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if *svsf.OplPathFlag == "" {
		if err := oplm.Create(sourceRoot, targetRoot); err != nil {
			common.Fatal(lgr, err)
		}
	}
	owo := &config.OpeLogOptionsType{
		SyncOptionsType: swo,
		Goal:            *oplCmdFlag,
		NoInvCheck:      *oplNoInvCFlag,
		SyncPeriod:      int64(syncPer),
		ResetMilTimeout: *oplRtoFlag,
		EnableRestart:   *oplRestartFlag,
		ClearErrors:     *oplClErrFlag,
		KeepCached:      *oplKeepCFlag,
		ExpFile:         *svsf.OplCsvFlag,
		ExpRpt:          *svsf.OplRptFlag,
		ExpPeriod:       int64(expPer),
		ExpBck:          expBck,
		ErrorRate:       *oplErrRtlFlag,
	}
	opti := &opelog.ToolInformation{
		CmdLine: strings.Join(os.Args, " "),
		Version: config.GetVersion(),
	}
	ow := opelogimpl.NewOplWalker(
		lgr, *cf.ConcurrencyFlag, nil, oplm,
		opti, owo,
		sDss, tDss,
		sourceRoot, targetRoot,
		*svsf.OplSesFlag, *svsf.OplInvFlag,
	)
	sh.RegisterCleanupFunction(func(lgr *slog.Logger) error {
		return ow.Shutdown(lgr)
	})
	if err = ow.Run(); err != nil {
		common.Fatal(lgr, err)
	}
	if *svsf.OplPathFlag == "" {
		if _, err := fmt.Fprintf(outFile, "%s\n", ow.ToolInfo().ExecSummary); err != nil {
			common.Fatal(lgr, err)
		}
	}
}
