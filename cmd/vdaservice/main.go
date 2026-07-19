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
)

func main() {
	var (
		cmdFlag   = flag.String("cmd", "", "a command to apply: list latency")
		dssFlag   = flag.String("dss", "", "dss on which the command applies")
		recurFlag = flag.Bool("recur", false, "apply recursively to sub-directories")
		sortFlag  = flag.Bool("sort", false, "sort output with entries paths")
		tsortFlag = flag.Bool("tsort", false, "sort output with entries modification times")
		noownFlag = flag.Bool("noown", false, "hide uid gid information")
		statFlag = flag.Bool("stat", false, "with list cmd, perform additional stat on each entry (simulate I/O)")
		latencyFlag   = flag.String("latency", "100us", "latency")
		countFlag   = flag.Int("count", 100000, "test count")
	)
	cf := cli.CommonFlags()
	svsf := cli.ServicesFlags()
	flag.Parse()
	lgr, err := common.CliLogger("vdasync", *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, err)
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

	if *dssFlag == "" {
		common.Fatal(lgr, errors.New("dss must be provided"))
	}

	dss, root, err := cli.GetDssAndRootFor(lgr, cf, cfg, false, *dssFlag, rps)
	_ = root
	if err != nil {
		common.Fatal(lgr, err)
	}
	defer dss.EndSession()

	if *cmdFlag == "" {
		common.Fatal(lgr, errors.New("cmd must be provided"))
	}
	err = cli.DoService(&cli.ServiceCtx{
		Cmd:         *cmdFlag,
		Dss:         dss,
		Root:        root,
		IsRecur:     *recurFlag,
		IsCheck:     *svsf.CheckFlag,
		IsStat:     *statFlag,
		CsAlgos:     *svsf.CsalFlag,
		IsSorted:    *sortFlag,
		IsTSorted:   *tsortFlag,
		IsNoOwn:     *noownFlag,
		Latency: *latencyFlag,
		Count: *countFlag,
		Concurrency: *cf.ConcurrencyFlag,
		Lgr:         lgr,
		OutFile:     outFile,
	})
	if err != nil {
		common.Fatal(lgr, err)
	}
	time.Sleep(10 * time.Millisecond)
}
