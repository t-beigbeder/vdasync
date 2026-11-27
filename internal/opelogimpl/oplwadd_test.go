package opelogimpl

import (
	"log/slog"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/walker"
	"github.com/t-beigbeder/vdasync/opelog"
)

func makeLocTestDss() (sds, tds dssa.Dssa, err error) {
	sds = localfiles.MakeLocalFilesDssa()
	if err = sds.NewSession(); err != nil {
		return
	}
	tds = localfiles.MakeLocalFilesDssa()
	if err = tds.NewSession(); err != nil {
		return
	}
	return
}
func TestSingleOplWalkers(t *testing.T) {
	t.Skip("for setup only")
	cliLgr, _ := common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	_ = cliLgr
	owts := []owTest{
		{
			label: "update once - new session",
			ftgen: ftGenAugTiny,
			conc:  4,
			owo: &config.OpeLogOptionsType{
				ExpRpt:          "full",
				SyncOptionsType: config.SyncOptionsType{Check: true},
			},
			allowReadPhase2: true,
			allowReadPhase4: true,
			session2:        "ses2",
			custoTest:       custTestNewSession,
		},
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
	require.NoError(t, ftGenAugSmall(std, nil))
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

	require.NoError(t, ftGenTiny(std, nil))
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

func TestAutoSessionSimple(t *testing.T) {
	var (
		lgr *slog.Logger
		err error
	)
	lgr = common.GetLogger()
	// lgr = common.InfoLogger()
	lgr, err = common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	lgr = common.DbgLogger()
	require.NoError(t, err)

	oplm, err := MakeM2fManager(lgr, "", 0)
	require.NoError(t, err)
	std, ttd := t.TempDir(), t.TempDir()
	defer walker.SetTestDirRW(std, "source")
	defer walker.SetTestDirRW(ttd, "target")
	require.NoError(t, ftGenAugTiny(std, nil))
	err = oplm.Create(std, ttd)
	require.NoError(t, err)
	require.Equal(t, "mem", oplm.GetOlm().StoreType)
	sds, tds, err := makeLocTestDss()
	require.NoError(t, err)
	ctd := path.Join(os.TempDir(), "oplCsv")
	os.Mkdir(ctd, 0700)
	csvPath := path.Join(ctd, "csv-test-auto-simple.csv")
	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&opelog.ToolInformation{CmdLine: "TestAutoSessionSimple"},
		&config.OpeLogOptionsType{
			SyncOptionsType: config.SyncOptionsType{
				Check: true,
			},
			Goal:            "update",
			NoInvCheck:      true,
			ResetMilTimeout: 250,
			ExpFile:         csvPath,
			ExpPeriod:       86400,
		},
		sds, tds, std, ttd, "", "",
	)
	err = ow.Run()
	require.NoError(t, err)
}

func TestAutoSessionDrCheck(t *testing.T) {
	var (
		lgr *slog.Logger
		err error
	)
	lgr = common.GetLogger()
	// lgr = common.InfoLogger()
	// lgr, err = common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	// lgr = common.DbgLogger()
	require.NoError(t, err)

	std, ttd := t.TempDir(), t.TempDir()
	defer walker.SetTestDirRW(std, "source")
	defer walker.SetTestDirRW(ttd, "target")
	require.NoError(t, ftGenAugTiny(std, nil))

	oplm, err := MakeM2fManager(lgr, "", 0)
	require.NoError(t, err)
	require.NoError(t, oplm.Create(std, ttd))
	require.Equal(t, "mem", oplm.GetOlm().StoreType)

	sds, tds, err := makeLocTestDss()
	require.NoError(t, err)
	ctd := path.Join(os.TempDir(), "oplCsv")
	os.Mkdir(ctd, 0700)
	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&opelog.ToolInformation{CmdLine: "TestAutoSessionDrCheck1"},
		&config.OpeLogOptionsType{
			Goal:            "update",
			NoInvCheck:      true,
			ResetMilTimeout: 250,
			ExpFile:         path.Join(ctd, "csv-test-auto-check1.csv"),
			ExpPeriod:       86400,
		},
		sds, tds, std, ttd, "", "",
	)
	require.NoError(t, ow.Run())

	oplm, err = MakeM2fManager(lgr, "", 0)
	require.NoError(t, err)
	require.NoError(t, oplm.Create(std, ttd))
	require.Equal(t, "mem", oplm.GetOlm().StoreType)

	ow = NewOplWalker(
		lgr, 4, nil, oplm,
		&opelog.ToolInformation{CmdLine: "TestAutoSessionDrCheck2"},
		&config.OpeLogOptionsType{
			SyncOptionsType: config.SyncOptionsType{
				Check:  true,
				Dryrun:  true,
			},
			Goal:            "update",
			NoInvCheck:      true,
			ResetMilTimeout: 250,
			ExpFile:         path.Join(ctd, "csv-test-auto-check2.csv"),
			ExpPeriod:       86400,
		},
		sds, tds, std, ttd, "", "",
	)
	require.NoError(t, ow.Run())
	require.NoError(t, oplm.Open(true))
	defer oplm.Close()
	olm := oplm.GetOlm()
	ases, ok := olm.Sessions["auto"]
	require.True(t, ok)
	rle, err := oplm.GetLogicalEntry("")
	require.NoError(t, err)
	require.NotNil(t, rle)
	cStats := rle.GetStats(ases.ReferenceTime)
	require.NotNil(t, cStats)
	require.Zero(t, cStats.Error.Number)
}
