package opelogimpl

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/walker"
	"github.com/t-beigbeder/vdasync/opelog"
)

type owTest struct {
	t                  *testing.T
	label              string
	unSkipped          bool
	lgr                *slog.Logger
	ftgen              func(root string) error
	conc               int
	owo                *config.OpeLogOptionsType
	oplm               opelog.OpeLogManager
	oplq               opelog.Queue
	hasLoadInv         bool
	hasNoExpCsv        bool
	csvName            string
	std, ltd, ttd, ctd string
	noTestMarsh        bool
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

var ftGenAugSmall = func(root string) error {
	_, _, err := walker.PrepareAugmentedTestFilesTree(root, 7, 100, 16, 1024)
	return err
}

func (owt *owTest) invImport() error {
	if !owt.hasLoadInv {
		return nil
	}
	cPath := path.Join(owt.ltd, "invDump.csv")
	if err := InventoryCsvExport(owt.std, cPath, owt.owo.InvCsAlgos); err != nil {
		return err
	}
	err := owt.oplm.Open(false)
	if err != nil {
		return err
	}
	defer owt.oplm.Close()
	iSs, ok := owt.oplm.Olm().Sessions["di"]
	if !ok {
		return errors.New("missing di session")
	}
	ica := owt.owo.InvCsAlgos
	if err := InventoryCsvImport(owt.oplm, iSs.ReferenceTime, cPath, ica); err != nil {
		return err
	}
	if err := owt.oplm.Close(); err != nil {
		return err
	}
	return nil
}

func (owt *owTest) exportCsv() error {
	if owt.hasNoExpCsv {
		return nil
	}
	csvName := strings.Join(strings.Split(owt.label, " "), "-")
	cPath := fmt.Sprintf("%s.csv", path.Join(owt.ctd, csvName))
	if err := owt.oplm.Open(true); err != nil {
		return err
	}
	defer owt.oplm.Close()
	if err := OplCsvExport(owt.lgr, owt.oplm, cPath, RPT_SYNTHETIC, "di", "ds"); err != nil {
		return err
	}
	owt.lgr.Info("exportCsv", "cPath", cPath)
	return nil
}

func (owt *owTest) setDefLogger() error {
	dbgLgr := common.DbgLogger()
	cliLgr, err := common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	infLgr := common.InfoLogger()
	defLgr := common.GetLogger()
	_, _, _, _ = dbgLgr, cliLgr, infLgr, defLgr
	defLgr = infLgr
	if err != nil {
		return err
	}
	if owt.lgr == nil {
		owt.lgr = defLgr
	}
	return nil
}

func (owt *owTest) runOwTest() (err error) {
	ctx := "ns"
	defer func() {
		if err == nil {
			return
		}
		err = fmt.Errorf("test \"%s\" - ctx %s - err %s", owt.label, ctx, err)
	}()
	if owt.ctd == "" {
		owt.ctd = path.Join(os.TempDir(), "oplCsv")
		os.Mkdir(owt.ctd, 0700)
	}
	owt.std = owt.t.TempDir()
	defer walker.SetTestDirRW(owt.std, "source")
	owt.ltd = owt.t.TempDir()
	owt.ttd = owt.t.TempDir()
	defer walker.SetTestDirRW(owt.ttd, "target")
	if owt.ftgen == nil {
		owt.ftgen = ftGenSmall
	}
	if owt.oplm == nil {
		if owt.oplm, err = makeTestM2fManager(path.Join(owt.ltd, "m2f.opl"), 0, !owt.noTestMarsh, owt.std, owt.ttd); err != nil {
			ctx = "makeTestM2fManager"
			return
		}
		if err = owt.oplm.Close(); err != nil {
			ctx = "oplmClose1"
			return
		}
	}
	if err = owt.ftgen(owt.std); err != nil {
		ctx = "ftgen"
		return
	}
	if err = owt.invImport(); err != nil {
		ctx = "invImport"
		return
	}
	if owt.owo.Goals == "" {
		owt.owo.Goals = "update"
	}
	owt.owo.ResetTimeout = -1
	ow := NewOplWalker(
		owt.lgr.With("test", owt.label), owt.conc,
		nil, owt.oplm, owt.owo,
		localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(),
		owt.std, owt.ttd,
		"ds", "di",
	)
	if err = ow.Run(); err != nil {
		ctx = "owRun"
		return
	}
	if err = owt.oplm.Open(true); err != nil {
		ctx = "oplmOpen"
		return
	}
	sTs := owt.oplm.Olm().Sessions["ds"].ReferenceTime
	rLe, err := owt.oplm.GetLogicalEntry("")
	if err != nil || rLe == nil || rLe.GetStats(sTs) == nil {
		err = errors.New("no stats on root")
		return
	}
	if nbe := rLe.GetStats(sTs).Error.Number; nbe != 0 {
		err = fmt.Errorf("got %d errors", nbe)
		return
	}
	if err = owt.oplm.Close(); err != nil {
		ctx = "oplmClose2"
		return
	}
	if err = owt.exportCsv(); err != nil {
		ctx = "exportCsv"
		return
	}
	return
}

func TestManyOplWalkers(t *testing.T) {

	owts := []owTest{
		{
			label:     "load - c0 small simple",
			unSkipped: false,
			lgr:       nil,
			ftgen:     ftGenSmall,
			conc:      0,
			owo:       &config.OpeLogOptionsType{
				// Goals: "update",
			},
			oplm:        nil,
			oplq:        nil,
			hasLoadInv:  false,
			noTestMarsh: false,
		},
		{
			label: "load c4 - medium with sync",
			conc:  4,
			ftgen: ftGenMedium,
			owo: &config.OpeLogOptionsType{
				Goals:      "load",
				SyncPeriod: 5,
			},
		},
		{
			label: "load and inv check - c4 small simple",
			ftgen: ftGenSmall,
			conc:  4,
			owo: &config.OpeLogOptionsType{
				Goals:      "load",
				InvCsAlgos: "md5",
			},
			hasLoadInv: true,
		},
		{
			label: "create - c4 tiny simple",
			ftgen: ftGenTiny,
			conc:  4,
			owo: &config.OpeLogOptionsType{
				Goals: "create",
			},
		},
		{
			label:       "update - c4 tiny simple",
			ftgen:       ftGenTiny,
			conc:        4,
			owo:         &config.OpeLogOptionsType{},
			noTestMarsh: true,
		},
		{
			label:       "update - c4 aug small",
			ftgen:       ftGenAugSmall,
			conc:        4,
			owo:         &config.OpeLogOptionsType{},
			noTestMarsh: true,
		},
	}

	skipDefault := false
	for _, owt := range owts {
		require.NoError(t, owt.setDefLogger())
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
	lgr = common.DbgLogger()
	require.NoError(t, err)

	lgr.Debug("TestOplWalker: started")
	std := t.TempDir()
	ltd := t.TempDir()
	ttd := t.TempDir()

	// require.NoError(t, ftGenTiny(std))
	// require.NoError(t, ftGenMedium(std))
	require.NoError(t, ftGenAugSmall(std))
	lgr.Debug("TestOplWalker: FileTreeGenerated")

	cPath := path.Join(ltd, "invDump.csv")
	require.NoError(t, InventoryCsvExport(std, cPath, "md5"))
	oplm, err := makeTestM2fManager(path.Join(ltd, "m2f.opl"), 5, true, std, ttd)
	require.NoError(t, err)
	require.Contains(t, oplm.Olm().Sessions, "di")
	iTs := oplm.Olm().Sessions["di"].ReferenceTime
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5"))
	require.NoError(t, oplm.Close())

	csvPath := path.Join(ltd, "oplm.csv")
	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&config.OpeLogOptionsType{
			SyncOptionsType: config.SyncOptionsType{
				Check: true,
			},
			Goals:        "update", 
			SyncPeriod:   5,
			ResetTimeout: -1,
			ExpPeriod:    6,
			ExpFile:      csvPath,
			InvCsAlgos:   "md5",
			NoInvCheck:   false,
		},
		localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(),
		std, ttd,
		"ds", "di")
	err = ow.Run()
	require.NoError(t, err)

	oplmRo, err := openTestM2fManagerRO(path.Join(ltd, "m2f.opl"), true)
	require.NoError(t, err)
	require.Contains(t, oplmRo.Olm().Sessions, "ds")
	sTs := oplm.Olm().Sessions["ds"].ReferenceTime
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
	oplm, err := makeTestM2fManager(path.Join(ltd, "m2f.opl"), 5, false, std, ttd)
	require.NoError(t, err)
	require.Contains(t, oplm.Olm().Sessions, "ds")
	iTs := oplm.Olm().Sessions["ds"].ReferenceTime
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5"))
	require.NoError(t, oplm.Close())
	require.Contains(t, oplm.Olm().Sessions, "di")
	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&config.OpeLogOptionsType{Goals: "load"},
		localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(),
		std, ttd,
		"di", "ds")
	err = ow.Run()
	require.Error(t, err)

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
}
