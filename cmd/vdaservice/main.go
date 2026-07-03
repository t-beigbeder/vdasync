package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/plugin"
)

type serviceCtx struct {
	cmd     string
	dss     dssa.Dssa
	root    string
	isRecur bool
	isCheck bool
	cf      *cli.CommonFlagsType
	lgr     *slog.Logger
	outFile io.WriteCloser
}

func doService(sc *serviceCtx) error {
	if sc.cmd == "list" {
		return doList(sc)
	}
	return fmt.Errorf("unknown command: %s", sc.cmd)
}

func doList(sc *serviceCtx) error {
	if !sc.isRecur {
		des, err := sc.dss.List(sc.root)
		if err != nil {
			return err
		}
		for _, de := range des {
			sc.outFile.Write([]byte(fmt.Sprintf("%+v\n", de)))
		}
		return nil
	}
	return fmt.Errorf("not implemented %+v", sc)
}

func main() {
	var (
		cmdFlag   = flag.String("cmd", "", "a command to apply: list")
		dssFlag   = flag.String("dss", "", "dss on which the command applies")
		checkFlag = flag.Bool("check", false, "compute/display checksums")
		recurFlag = flag.Bool("recur", false, "apply recursively to sub-directories")
		exclFlag  = flag.String("excl", "", "file containing regexps for paths to be excluded, defaults to none")
		inclFlag  = flag.String("incl", "", "file containing regexps for paths to be included, defaults to all")
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
	err = doService(&serviceCtx{
		cmd:     *cmdFlag,
		dss:     dss,
		root:    root,
		isRecur: *recurFlag,
		isCheck: *checkFlag,
		cf:      cf,
		lgr:     lgr,
		outFile: outFile,
	})
	if err != nil {
		common.Fatal(lgr, err)
	}
	time.Sleep(10 * time.Millisecond)
}
