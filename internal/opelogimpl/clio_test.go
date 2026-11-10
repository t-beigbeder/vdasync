package opelogimpl

import (
	"log/slog"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/walker"
	"github.com/t-beigbeder/vdasync/opelog"
)

func TestCliOutput(t *testing.T) {
	var (
		lgr *slog.Logger
		err error
	)
	lgr = common.GetLogger()
	// lgr = common.InfoLogger()
	// lgr, err = common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	// lgr = common.DbgLogger()
	require.NoError(t, err)

	lgr.Debug("TestCliOutput: started")
	std := t.TempDir()
	defer walker.SetTestDirRW(std, "source")
	ltd := t.TempDir()
	ttd := t.TempDir()
	defer walker.SetTestDirRW(ttd, "target")

	// require.NoError(t, ftGenTiny(std))
	// require.NoError(t, ftGenMedium(std))
	require.NoError(t, ftGenAugSmall(std))
	lgr.Debug("TestCliOutput: FileTreeGenerated")
	testMarsh := true
	oplm, err := makeTestM2fManager(lgr, path.Join(ltd, "m2f.opl"), 5, testMarsh, std, ttd)
	require.NoError(t, err)
	require.NoError(t, oplm.Close())

	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&opelog.ToolInformation{CmdLine: "TestCliOutput"},
		&config.OpeLogOptionsType{
			SyncOptionsType: config.SyncOptionsType{
				Check: true,
			},
			Goal:            "update",
			ResetMilTimeout: 12,
			NoInvCheck:      true,
		},
		localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(),
		std, ttd,
		"ds", "")
	err = ow.Run()
	require.NoError(t, err)

	oplmRo, err := openTestM2fManagerRO(lgr, path.Join(ltd, "m2f.opl"), testMarsh)
	require.NoError(t, err)
	require.Contains(t, oplmRo.GetOlm().Sessions, "ds")
	sTs := oplm.GetOlm().Sessions["ds"].ReferenceTime
	rLe, err := oplmRo.GetLogicalEntry("")
	require.NoError(t, err)
	require.Zero(t, rLe.GetStats(sTs).Error.Number)

	outPath := path.Join(os.TempDir(), "cliout.txt")
	out, err := os.Create(outPath)
	require.NoError(t, err)
	require.NoError(t, CliOutput(lgr, oplmRo, out, "ds", true, true))
}
