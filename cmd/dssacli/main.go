package main

import (
	"errors"
	"flag"
	"os"

	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/internal/cli"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/otvl_dtacsy/internal/walker"
)

func main() {
	var (
		sourceFlag      = flag.String("source", "", "source of the command if sync")
		targetFlag      = flag.String("target", "", "target of the command")
		dryRunFlag      = flag.Bool("dryrun", false, "don't run operation, just report actions")
		rmFlag          = flag.Bool("rm", false, "remove files in sync target")
	)
	cf := cli.CommonFlags()
	flag.Parse()
	lgr, err := common.CliLogger("dssacli", *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if *sourceFlag == "" || *targetFlag == "" {
		common.Fatal(lgr, errors.New("source and target must be provided"))
	}
	dss := localfiles.MakeLocalFilesDssa()
	so := &config.SyncOptionsType{}
	so.Dryrun = *dryRunFlag
	so.Rm = *rmFlag
	swk := walker.NewSynchronizer(lgr, *cf.ConcurrencyFlag, so, dss, dss, *targetFlag)
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
