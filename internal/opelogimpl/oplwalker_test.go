package opelogimpl

import (
	"log/slog"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/opelog"
)

type owTest struct {
	label         string
	unSkipped     bool
	lgr           *slog.Logger
	ftgen         func(root string) error
	conc          int
	owo           *config.OpeLogOptionsType
	oplm          opelog.OpeLogManager
	oplq          opelog.Queue
	loadInv       bool
	std, ltd, ttd string
}

var ftGenTiny = func(root string) error {
	return common.FileTreeGenerate(root, 5, 20, 1, 1024, false, 2)
}

var ftGenSmall = func(root string) error {
	return common.FileTreeGenerate(root, 20, 600, 2, 1024, false, 2)
}

var ftGenMedium = func(root string) error {
	return common.FileTreeGenerate(root, 100, 3000, 2, 4096, false, 2)
}

func (owt *owTest) invImport() error {
	if !owt.loadInv {
		return nil
	}
	cPath := path.Join(owt.ltd, "invDump.csv")
	if err := InventoryCsvExport(owt.std, cPath, owt.owo.InvCsAlgos); err != nil {
		return err
	}
	if err := owt.oplm.Open(false); err != nil {
		return err
	}
	defer owt.oplm.Close()
	if err := InventoryCsvImport(owt.oplm, cPath, owt.owo.InvCsAlgos); err != nil {
		return err
	}
	if err := owt.oplm.Close(); err != nil {
		return err
	}
	return nil
}

func (owt *owTest) invCheck() error {
	if !owt.loadInv {
		return nil
	}
	cPath := path.Join(owt.ltd, "oplmExport.csv")
	if err := owt.oplm.Open(true); err != nil {
		return err
	}
	defer owt.oplm.Close()
	if err := OplCsvExport(owt.oplm, cPath, RPT_SYNTHETIC); err != nil {
		return err
	}
	owt.lgr.Info("invCheck", "csvExport", cPath)
	return nil
}

func TestManyOplWalkers(t *testing.T) {
	var (
		err error
	)
	dbgLgr := common.DbgLogger()
	cliLgr, err := common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	require.NoError(t, err)
	infLgr := common.InfoLogger()
	defLgr := common.GetLogger()
	_, _, _, _ = dbgLgr, cliLgr, infLgr, defLgr
	skipDefault := false
	defLgr = infLgr

	owts := []owTest{
		{
			label:     "load - c0 small simple",
			unSkipped: false,
			lgr:       defLgr,
			ftgen:     ftGenSmall,
			conc:      0,
			owo: &config.OpeLogOptionsType{
				Goals: "load", // load, create, update/remove, verify
			},
			oplm:    nil,
			oplq:    nil,
			loadInv: false,
		},
		{
			label: "load c4 - medium with sync",
			conc:  4,
			ftgen: ftGenMedium,
			owo: &config.OpeLogOptionsType{
				Goals:      "load",
				SyncPeriod: int64(5 * time.Second),
			},
		},
		{
			label:     "load & inv check - c4 small simple",
			ftgen:     ftGenSmall,
			unSkipped: true,
			conc:      4,
			owo: &config.OpeLogOptionsType{
				Goals:      "load",
				InvCsAlgos: "md5",
			},
			loadInv: true,
		},
		{
			label:     "create - c4 tiny simple",
			unSkipped: true,
			ftgen:     ftGenTiny,
			conc:      4,
			owo: &config.OpeLogOptionsType{
				Goals: "create",
			},
		},
	}

	for _, owt := range owts {
		if owt.lgr == nil {
			owt.lgr = defLgr
		}
		if skipDefault {
			if !owt.unSkipped {
				owt.lgr.Info("skipped", "test", owt.label)
				continue
			}
		}
		owt.std = t.TempDir()
		owt.ltd = t.TempDir()
		owt.ttd = t.TempDir()
		if owt.ftgen == nil {
			owt.ftgen = ftGenSmall
		}
		if owt.oplm == nil {
			owt.oplm, err = MakeM2fManager(path.Join(owt.ltd, "m2f.opl"))
		}
		require.NoError(t, owt.oplm.Create(owt.std, owt.ttd))
		require.NoError(t, owt.ftgen(owt.std))
		require.NoError(t, owt.invImport())
		ow := NewOplWalker(
			owt.lgr.With("test", owt.label), owt.conc,
			nil, owt.oplm, owt.owo,
			localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(),
			owt.std, owt.ttd)
		require.NoError(t, ow.Run())
		require.NoError(t, owt.invCheck())
		csvPath := path.Join(owt.ltd, "oplm.csv")
		require.NoError(t, owt.oplm.Open(true))
		require.NoError(t, OplCsvExport(owt.oplm, csvPath, RPT_SYNTHETIC))
		owt.lgr.Info("exported", "csv", csvPath)
		require.NoError(t, owt.oplm.Close())
		require.True(t, true)
	}
}

func TestOplWalker(t *testing.T) {
	//t.Skip("wip")
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
	require.NoError(t, common.FileTreeGenerate(std, 100, 3000, 2, 4096, false, 2))
	lgr.Debug("TestOplWalker: FileTreeGenerated")

	ltd := t.TempDir()
	ttd := t.TempDir()

	oplm, err := MakeM2fManager(path.Join(ltd, "m2f.opl"))
	require.NoError(t, err)
	require.NoError(t, oplm.Create(std, ttd))

	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&config.OpeLogOptionsType{
			Goals:      "load", // load, create, update/remove, verify
			SyncPeriod: int64(5 * time.Second),
		},
		localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(), std, ttd)
	err = ow.Run()
	require.NoError(t, err)
}

func TestGoals(t *testing.T) {
	owi := &oplWalkerImpl{owo: &config.OpeLogOptionsType{Goals: ""}}
	owi.owo.Goals = "load"
	require.True(t, owi.impliesGoal("load"))
	require.False(t, owi.impliesGoal("create"))
	owi.owo.Goals = "create"
	require.True(t, owi.impliesGoal("load"))
	require.True(t, owi.impliesGoal("create"))
	require.False(t, owi.impliesGoal("update"))
	owi.owo.Goals = "update"
	require.True(t, owi.impliesGoal("load"))
	require.True(t, owi.impliesGoal("create"))
	require.True(t, owi.impliesGoal("update"))
	require.False(t, owi.impliesGoal("verify"))
	owi.owo.Goals = "load,create,update,verify2"
	require.True(t, owi.impliesGoal("load"))
	require.True(t, owi.impliesGoal("create"))
	require.True(t, owi.impliesGoal("update"))
	require.False(t, owi.impliesGoal("verify"))
	owi.owo.Goals = "load,create,update,verify"
	require.True(t, owi.impliesGoal("load"))
	require.True(t, owi.impliesGoal("create"))
	require.True(t, owi.impliesGoal("update"))
	require.True(t, owi.impliesGoal("verify"))
	owi.owo.Goals = "verify"
	require.False(t, owi.impliesGoal("load"))
	require.False(t, owi.impliesGoal("create"))
	require.False(t, owi.impliesGoal("update"))
	require.True(t, owi.impliesGoal("verify"))
}
