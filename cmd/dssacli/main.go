package main

import (
	"errors"
	"flag"
	"os"

	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/otvl_dtacsy/internal/walker"
)

func main() {
	var (
		sourceFlag      = flag.String("source", "", "source of the command if sync")
		targetFlag      = flag.String("target", "", "target of the command")
		concurrencyFlag = flag.Int("conc", 0, "number of concurrent activities")
		dryRunFlag      = flag.Bool("dryrun", false, "don't run operation, just report actions")
		rmFlag          = flag.Bool("rm", false, "remove files in sync target")
		llFlag          = flag.String("level", "", "log level, defaults to ERROR")
		logFlag         = flag.String("log", "", "log file, defaults to dssacli-<pid>.log in temp dir, \"stderr\" is a known keyword")
	)
	lgr, err := common.CliLogger("dssacli", *llFlag, *logFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	flag.Parse()
	if *sourceFlag == "" || *targetFlag == "" {
		common.Fatal(lgr, errors.New("source and target must be provided"))
	}
	dss := localfiles.MakeLocalFilesDssa()
	so := &config.SyncOptionsType{}
	so.Dryrun = *dryRunFlag
	so.Rm = *rmFlag
	swk := walker.NewSynchronizer(lgr, *concurrencyFlag, so, dss, dss, *targetFlag)
	sde, err := dss.Stat(*sourceFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if err = swk.Run(sde); err != nil {
		common.Fatal(lgr, err)
	}
	syncRes := walker.SyncResult(swk)
	walker.DisplaySyncResult(syncRes, os.Stdout, true, false)
}
