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

func TestSingleOplWalkers(t *testing.T) {
	t.Skip("for setup only")
	cliLgr, _ := common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	_ = cliLgr
	owts := []owTest{
		{
			label: "test - update once - new session as load",
			lgr: common.DbgLogger(),
			ftgen: ftGenAugSmall,
			owo: &config.OpeLogOptionsType{
				SyncOptionsType: config.SyncOptionsType{Check: true},
			},
			session2:        "ses2",
			custoTest:       custTestNewSession,
			ctDict: map[string]string{"ns-goal": "load"},
			specialInterest: []string{""},
			speIntPhase:     "phase3",
		},
		// {
		// 	label: "test - update once - inventory - no keep cache then again",
		// 	ftgen: ftGenTiny,
		// 	owo: &config.OpeLogOptionsType{
		// 		SyncOptionsType: config.SyncOptionsType{Check: true},
		// 		InvCsAlgos:      "md5",
		// 	},
		// 	hasLoadInv: true,
		// 	custoTest:  custTestNoKeepCached,
		// },
	}

	skipDefault := false
	for _, owt := range owts {
		if skipDefault {
			if !owt.unSkipped {
				owt.lgr.Info("skipped", "test", owt.label)
				continue
			}
		}
		owt.t = t
		err := owt.runOwTest()
		require.NoError(t, err, err)
	}
}

func TestOplWalker(t *testing.T) {
	var (
		lgr *slog.Logger
		err error
	)
	lgr = common.GetLogger()
	// lgr = common.InfoLogger()
	// lgr, err = common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	// lgr = common.DbgLogger()
	require.NoError(t, err)

	lgr.Debug("TestOplWalker: started")
	std := t.TempDir()
	defer walker.SetTestDirRW(std, "source")
	ltd := t.TempDir()
	ttd := t.TempDir()
	defer walker.SetTestDirRW(ttd, "target")

	// require.NoError(t, ftGenTiny(std))
	// require.NoError(t, ftGenMedium(std))
	require.NoError(t, ftGenAugSmall(std))
	lgr.Debug("TestOplWalker: FileTreeGenerated")

	cPath := path.Join(ltd, "invDump.csv")
	require.NoError(t, InventoryCsvExport(std, cPath, "md5"))
	oplm, err := makeTestM2fManager(lgr, path.Join(ltd, "m2f.opl"), 5, true, std, ttd)
	require.NoError(t, err)
	require.Contains(t, oplm.GetOlm().Sessions, "di")
	iTs := oplm.GetOlm().Sessions["di"].ReferenceTime
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5"))
	require.NoError(t, oplm.Close())

	csvPath := path.Join(ltd, "oplm.csv")
	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&opelog.ToolInformation{CmdLine: "TestOplWalker"},
		&config.OpeLogOptionsType{
			SyncOptionsType: config.SyncOptionsType{
				Check: true,
			},
			Goal:            "update",
			SyncPeriod:      5,
			ResetMilTimeout: 250,
			ExpPeriod:       6,
			ExpFile:         csvPath,
			InvCsAlgos:      "md5",
			NoInvCheck:      false,
		},
		localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(),
		std, ttd,
		"ds", "di")
	err = ow.Run()
	require.NoError(t, err)

	oplmRo, err := openTestM2fManagerRO(lgr, path.Join(ltd, "m2f.opl"), true)
	require.NoError(t, err)
	require.Contains(t, oplmRo.GetOlm().Sessions, "ds")
	sTs := oplm.GetOlm().Sessions["ds"].ReferenceTime
	rLe, err := oplmRo.GetLogicalEntry("")
	require.NoError(t, err)
	require.Zero(t, rLe.GetStats(sTs).Error.Number)
}

func TestOplBadSession(t *testing.T) {
	var (
		lgr *slog.Logger
		err error
	)
	lgr = common.GetLogger()
	// lgr = common.InfoLogger()
	// lgr, err = common.CliLogger("TestOplBadSession", "DEBUG+2", "stderr")
	// lgr = common.DbgLogger()
	require.NoError(t, err)

	lgr.Debug("TestOplBadSession: started")
	std := t.TempDir()
	ltd := t.TempDir()
	ttd := t.TempDir()

	require.NoError(t, ftGenTiny(std))
	lgr.Debug("TestOplBadSession: FileTreeGenerated")

	cPath := path.Join(ltd, "invDump.csv")
	require.NoError(t, InventoryCsvExport(std, cPath, "md5"))
	oplm, err := makeTestM2fManager(lgr, path.Join(ltd, "m2f.opl"), 5, false, std, ttd)
	require.NoError(t, err)
	require.Contains(t, oplm.GetOlm().Sessions, "ds")
	iTs := oplm.GetOlm().Sessions["ds"].ReferenceTime
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5"))
	require.NoError(t, oplm.Close())
	require.Contains(t, oplm.GetOlm().Sessions, "di")
	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&opelog.ToolInformation{CmdLine: "TestOplWalker"},
		&config.OpeLogOptionsType{Goal: "load"},
		localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(),
		std, ttd,
		"di", "ds")
	err = ow.Run()
	require.Error(t, err)

}

func TestGoals(t *testing.T) {
	owi := &oplWalkerImpl{owo: &config.OpeLogOptionsType{Goal: ""}}
	owi.owo.Goal = "load"
	require.True(t, owi.impliesGoal("load"))
	require.False(t, owi.impliesGoal("create"))
	owi.owo.Goal = "create"
	require.True(t, owi.impliesGoal("load"))
	require.True(t, owi.impliesGoal("create"))
	require.False(t, owi.impliesGoal("update"))
	owi.owo.Goal = "update"
	require.True(t, owi.impliesGoal("load"))
	require.True(t, owi.impliesGoal("create"))
	require.True(t, owi.impliesGoal("update"))
}

func TestAutoSession(t *testing.T) {
	var (
		lgr *slog.Logger
		err error
	)
	lgr = common.GetLogger()
	// lgr = common.InfoLogger()
	// lgr, err = common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	// lgr = common.DbgLogger()
	require.NoError(t, err)

	oplm, err := MakeM2fManager(lgr, "", 0)
	require.NoError(t, err)
	std, ttd := t.TempDir(), t.TempDir()
	defer walker.SetTestDirRW(std, "source")
	defer walker.SetTestDirRW(ttd, "target")
	require.NoError(t, ftGenAugTiny(std))
	err = oplm.Create(std, ttd)
	require.NoError(t, err)
	require.Equal(t, "mem", oplm.GetOlm().StoreType)
	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&opelog.ToolInformation{CmdLine: "TestAutoSession"},
		&config.OpeLogOptionsType{
			SyncOptionsType: config.SyncOptionsType{
				Check: true,
			},
			Goal:            "update",
			NoInvCheck:      true,
			ResetMilTimeout: 250,
		},
		localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(),
		std, ttd, "", "",
	)
	err = ow.Run()
	require.NoError(t, err)
	ctd := path.Join(os.TempDir(), "oplCsv")
	csvPath := path.Join(ctd, "csv-test-auto.csv")
	oplm.Open(true)
	err = OplCsvExport(lgr, oplm, csvPath, RPT_SYNTHETIC, "", "auto", nil)
	require.NoError(t, err)
}
