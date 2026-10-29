package opelogimpl

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/proxy"
	"github.com/t-beigbeder/vdasync/internal/walker"
	"github.com/t-beigbeder/vdasync/opelog"
)

type owTest struct {
	t                  *testing.T
	label              string
	ctx                string
	unSkipped          bool
	lgr                *slog.Logger
	ftgen              func(root string) error
	ftupd              func(root string) error
	conc               int
	oplTi              *opelog.ToolInformation
	owo                *config.OpeLogOptionsType
	oplm               opelog.OpeLogManager
	oplq               opelog.Queue
	session1           string
	session2           string
	inventory          string
	hasLoadInv         bool
	hasNoExpCsv        bool
	hasNoRm            bool
	hasNoCheck         bool
	restartCb          func(phase string, count int) bool
	hasNoPhase2        bool
	hasNoExpCsv2       bool
	hasNoPhase4        bool
	hasNoExpCsv4       bool
	csvName            string
	std, ltd, ttd, ctd string
	noGenSDs, noGenTDs bool
	sDss, tDss         dssa.Dssa
	sIoeOpts           *proxy.IoErrOpts
	tIoeOpts           *proxy.IoErrOpts
	noTestMarsh        bool
	specialInterest    []string
	speIntPhase        string
	custoTest          func(*owTest) error
	deferred           func()
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

var ftGenAugTiny = func(root string) error {
	_, _, err := walker.PrepareAugmentedTestFilesTree(root, 5, 20, 16, 1024)
	return err
}

var ftUpdAugTiny = func(root string) error {
	_, _, err := walker.UpdateAugmentedTestFilesTree(root, 3, 7, 2, 1027)
	return err
}

var ftGenAugSmall = func(root string) error {
	_, _, err := walker.PrepareAugmentedTestFilesTree(root, 7, 100, 16, 1024)
	return err
}

var ftUpdAugSmall = func(root string) error {
	_, _, err := walker.UpdateAugmentedTestFilesTree(root, 5, 10, 3, 1027)
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
	iSs, ok := owt.oplm.GetOlm().Sessions[owt.inventory]
	if !ok {
		return fmt.Errorf("missing %s inventory", owt.inventory)
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

func (owt *owTest) makeOplm() (err error) {
	if owt.oplm == nil {
		if owt.oplm, err = MakeM2fManager(owt.lgr, path.Join(owt.ltd, "m2f.opl"), 0); err != nil {
			return
		}
		oplmi, ok := owt.oplm.(*m2fOplMng)
		if !ok {
			err = errors.New("not an *m2fOplMng")
			return
		}
		oplmi.testMarsh = !owt.noTestMarsh
		if err = owt.oplm.Create(owt.std, owt.ttd); err != nil {
			return
		}
		if err = owt.oplm.Open(false); err != nil {
			return
		}
		if _, err = owt.oplm.AddSession(owt.session1, &opelog.SessionInfo{}); err != nil {
			return
		}
		if _, err = owt.oplm.AddSession(owt.inventory, &opelog.SessionInfo{IsInventory: true}); err != nil {
			return
		}
		if owt.session2 != "" {
			if _, err = owt.oplm.AddSession(owt.session2, &opelog.SessionInfo{}); err != nil {
				return
			}
		}
		if err = owt.oplm.Close(); err != nil {
			return
		}
	}
	return
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

func (owt *owTest) commonSetup() (err error) {
	ctx := "ns"
	defer func() {
		if err == nil {
			return
		}
		err = fmt.Errorf("test \"%s\" - ctx %s - err %s", owt.label, ctx, err)
	}()

	if err = owt.setDefLogger(); err != nil {
		ctx = "setDefLogger"
		return
	}

	if owt.ctd == "" {
		owt.ctd = path.Join(os.TempDir(), "oplCsv")
		os.Mkdir(owt.ctd, 0700)
	}
	owt.std = owt.t.TempDir()
	owt.ltd = owt.t.TempDir()
	owt.ttd = owt.t.TempDir()
	owt.deferred = func() {
		defer walker.SetTestDirRW(owt.std, "source")
		defer walker.SetTestDirRW(owt.ttd, "target")
	}
	if owt.ftgen == nil {
		owt.ftgen = ftGenSmall
	}
	if owt.session1 == "" {
		owt.session1 = "ses1"
	}
	if owt.inventory == "" {
		owt.inventory = "inv"
	}
	if err = owt.makeOplm(); err != nil {
		ctx = "makeOplm"
		return
	}
	if err = owt.ftgen(owt.std); err != nil {
		ctx = "ftgen"
		return
	}
	if err = owt.invImport(); err != nil {
		ctx = "invImport"
		return
	}
	if owt.owo.Goal == "" {
		owt.owo.Goal = "update"
	}

	// WARN: not fully reliable, could become flaky on low resource dev env
	owt.owo.ResetMilTimeout = 250
	owt.owo.SpecialInterest = owt.specialInterest
	owt.owo.SpeIntPhase = owt.speIntPhase
	owt.owo.Check = !owt.hasNoCheck
	if !owt.hasNoCheck && !owt.hasLoadInv {
		owt.owo.NoInvCheck = true
	}

	return
}

func (owt *owTest) exportCsv(baseCtx string, session, inventory string) error {
	if owt.hasNoExpCsv {
		return nil
	}
	if (baseCtx == "phase2" && owt.hasNoExpCsv2) || (baseCtx == "phase4" && owt.hasNoExpCsv4) {
		return nil
	}
	csvName := strings.Join(strings.Split(owt.label, " "), "-")
	if baseCtx != "" {
		csvName += "-" + baseCtx
	}
	cPath := fmt.Sprintf("%s.csv", path.Join(owt.ctd, csvName))
	if err := owt.oplm.Open(true); err != nil {
		return err
	}
	defer owt.oplm.Close()
	rpt := RPT_SYNTHETIC
	if owt.owo.ExpRpt == "full" {
		rpt = RPT_FULL
	}
	if err := OplCsvExport(owt.lgr, owt.oplm, cPath, rpt, inventory, session, &RptConfig{ToolInfo: owt.oplTi, IsSorted: true}); err != nil {
		return err
	}
	owt.lgr.Info("exportCsv", "cPath", cPath)
	return nil
}

func (owt *owTest) newOplWalker(session, inventory string) OplWalker {
	if owt.sDss == nil || !owt.noGenSDs {
		owt.sDss = localfiles.MakeLocalFilesDssa()
		if owt.sIoeOpts != nil {
			owt.sDss = proxy.MakeIoErrDss(owt.lgr, owt.sDss, owt.sIoeOpts)
		}
	}
	if owt.tDss == nil || !owt.noGenTDs {
		owt.tDss = localfiles.MakeLocalFilesDssa()
		if owt.tIoeOpts != nil {
			owt.tDss = proxy.MakeIoErrDss(owt.lgr, owt.tDss, owt.tIoeOpts)
		}
	}
	cl := fmt.Sprintf("%s-%s", owt.t.Name(), owt.label)
	if owt.oplTi == nil {
		owt.oplTi = &opelog.ToolInformation{CmdLine: cl}
	}
	return NewOplWalker(
		owt.lgr.With("test", owt.label).With("ctx", owt.ctx), owt.conc,
		nil, owt.oplm,
		owt.oplTi, owt.owo,
		owt.sDss, owt.tDss, owt.std, owt.ttd,
		session, inventory,
	)
}

func (owt *owTest) walkStatsCsv(baseCtx string, session, inventory string) (ctx string, stats *opelog.ComputedStats, err error) {
	defer func() {
		if ctx != "" {
			ctx = fmt.Sprintf("%s: %s", baseCtx, ctx)
		}
	}()
	owt.ctx = baseCtx
	ow := owt.newOplWalker(session, inventory)
	if err = owt.sDss.NewSession(); err != nil {
		ctx = "sDssNewSession"
		return
	}
	if err = owt.tDss.NewSession(); err != nil {
		ctx = "tDssNewSession"
		return
	}
	owi := ow.(*oplWalkerImpl)
	owi.testCtx = baseCtx
	owi.restartCb = owt.restartCb
	if err = ow.Run(); err != nil {
		ctx = "owRun"
		return
	}
	if err = owt.oplm.Open(true); err != nil {
		ctx = "oplmOpen"
		return
	}
	sTs := owt.oplm.GetOlm().Sessions[session].ReferenceTime
	rLe, err := owt.oplm.GetLogicalEntry("")
	if err != nil || rLe == nil || rLe.GetStats(sTs) == nil {
		err = errors.New("no stats on root")
		ctx = "GetStats"
		return
	}
	stats = rLe.GetStats(sTs)
	if err = owt.oplm.Close(); err != nil {
		ctx = "oplmClose"
		return
	}
	if err = owt.exportCsv(baseCtx, session, inventory); err != nil {
		ctx = "exportCsv"
		return
	}
	if nbe := stats.Error.Number; nbe != 0 {
		ctx = "errors"
		err = fmt.Errorf("got %d errors", nbe)
		return
	}
	return
}

func (owt *owTest) runOwTest() (err error) {
	if owt.custoTest != nil {
		return owt.custoTest(owt)
	}
	var stats *opelog.ComputedStats
	if err = owt.commonSetup(); err != nil {
		return
	}
	ctx := "ns"
	defer func() {
		if owt.deferred != nil {
			owt.deferred()
		}
		if err == nil {
			return
		}
		err = fmt.Errorf("test \"%s\" - ctx %s - err %s", owt.label, ctx, err)
	}()

	// phase1: initial sync request
	if ctx, _, err = owt.walkStatsCsv("phase1", owt.session1, owt.inventory); err != nil {
		return
	}

	// phase2: dryrun
	if !owt.hasNoPhase2 {
		time.Sleep(time.Millisecond)
		owt.owo.Dryrun = true
		owt.owo.KeepCached = true // to compare w/o resetting state
		if ctx, stats, err = owt.walkStatsCsv("phase2", owt.session1, owt.inventory); err != nil {
			return
		}
		if stats.IsSet(true) {
			ctx = "phase2"
			err = errors.New("stats is set")
			return
		}
	}

	// phase3: update
	owt.owo.Check = !owt.hasNoCheck
	if !owt.hasNoCheck && !owt.hasLoadInv {
		owt.owo.NoInvCheck = true
	}
	if owt.ftupd != nil {
		time.Sleep(time.Millisecond)
		owt.owo.Dryrun = false
		owt.owo.KeepCached = false
		owt.owo.Rm = !owt.hasNoRm
		if err = owt.ftupd(owt.std); err != nil {
			ctx = "ftupd"
			return
		}
		if ctx, stats, err = owt.walkStatsCsv("phase3", owt.session1, owt.inventory); err != nil {
			return
		}
	}

	// phase4: dryrun for phase3 if applicable
	if owt.ftupd != nil && !owt.hasNoPhase4 {
		time.Sleep(time.Millisecond)
		owt.owo.Dryrun = true
		owt.owo.KeepCached = true
		if ctx, stats, err = owt.walkStatsCsv("phase4", owt.session1, owt.inventory); err != nil {
			return
		}
		if stats.IsSet(true) {
			ctx = "phase4"
			err = errors.New("stats is set")
		}
	}

	return

}

func custTestNoKeepCached(owt *owTest) (err error) {
	var stats *opelog.ComputedStats
	if err = owt.commonSetup(); err != nil {
		return
	}
	ctx := "ns"
	defer func() {
		if owt.deferred != nil {
			owt.deferred()
		}
		if err == nil {
			return
		}
		err = fmt.Errorf("test \"%s\" - ctx %s - err %s", owt.label, ctx, err)
	}()

	// phase1: initial sync request
	if ctx, _, err = owt.walkStatsCsv("phase1", owt.session1, owt.inventory); err != nil {
		return
	}

	// phase2: dryrun
	time.Sleep(time.Millisecond)
	owt.owo.Dryrun = true
	if ctx, stats, err = owt.walkStatsCsv("phase2", owt.session1, owt.inventory); err != nil {
		return
	}

	// phase1bis: again
	time.Sleep(time.Millisecond)
	owt.owo.Dryrun = false
	if ctx, stats, err = owt.walkStatsCsv("phase1bis", owt.session1, owt.inventory); err != nil {
		return
	}

	// phase2bis: dryrun again
	time.Sleep(time.Millisecond)
	owt.owo.Dryrun = true
	owt.owo.KeepCached = true
	if ctx, stats, err = owt.walkStatsCsv("phase2bis", owt.session1, owt.inventory); err != nil {
		return
	}
	if stats.IsSet(true) {
		ctx = "phase2bis"
		err = errors.New("stats is set")
	}

	return
}

func custTestNewSession(owt *owTest) (err error) {
	var stats *opelog.ComputedStats
	if err = owt.commonSetup(); err != nil {
		return
	}
	ctx := "ns"
	defer func() {
		if owt.deferred != nil {
			owt.deferred()
		}
		if err == nil {
			return
		}
		err = fmt.Errorf("test \"%s\" - ctx %s - err %s", owt.label, ctx, err)
	}()

	// phase1: initial sync request
	if ctx, _, err = owt.walkStatsCsv("phase1", owt.session1, owt.inventory); err != nil {
		return
	}

	// phase3: update
	if ctx, _, err = owt.walkStatsCsv("phase3", owt.session2, owt.inventory); err != nil {
		return
	}

	// phase4: dryrun for phase3 if applicable
	time.Sleep(time.Millisecond)
	owt.owo.Dryrun = true
	owt.owo.KeepCached = true
	if ctx, stats, err = owt.walkStatsCsv("phase4", owt.session1, owt.inventory); err != nil {
		return
	}
	if stats.IsSet(true) {
		ctx = "phase4"
		err = errors.New("stats is set")
	}

	return
}

func TestManyOplWalkers(t *testing.T) {

	owts := []owTest{
		{
			label:     "update - c0 small simple",
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
				Goal:       "load",
				SyncPeriod: 5,
			},
		},
		{
			label: "load and inv check - c4 small simple",
			ftgen: ftGenSmall,
			conc:  4,
			owo: &config.OpeLogOptionsType{
				Goal:       "load",
				InvCsAlgos: "md5",
			},
			hasLoadInv: true,
		},
		{
			label: "create - c4 tiny",
			ftgen: ftGenTiny,
			conc:  4,
			owo: &config.OpeLogOptionsType{
				Goal: "create",
			},
		},
		{
			label: "update - c4 tiny",
			ftgen: ftGenTiny,
			conc:  4,
			owo:   &config.OpeLogOptionsType{},
		},
		{
			label: "update - c4 aug small",
			ftgen: ftGenAugSmall,
			conc:  4,
			owo:   &config.OpeLogOptionsType{},
		},
		{
			label: "update - increment - c4 aug small",
			ftgen: ftGenAugSmall,
			ftupd: ftUpdAugSmall,
			conc:  4,
			owo:   &config.OpeLogOptionsType{},
		},
		{
			label: "update - read io errors",
			conc:  1,
			ftgen: ftGenTiny,
			owo: &config.OpeLogOptionsType{
				ErrorRate:     "200",
				ExpRpt:        "full",
				EnableRestart: true,
				ClearErrors:   true,
				KeepCached:    true,
			},
			sIoeOpts: &proxy.IoErrOpts{
				ResetOnNew: true,
				ListRate:   13, StatRate: 13, GetReadCloserRate: 13, ReadRate: 130, RCloseRate: 13,
			},
			restartCb: func(phase string, count int) bool {
				if phase != "phase1" || count >= 2 {
					return false
				}
				return true
			},
		},
		{
			label: "update - read write io errors",
			conc:  1,
			ftgen: ftGenAugSmall,
			owo: &config.OpeLogOptionsType{
				ErrorRate:     "200",
				ExpRpt:        "full",
				EnableRestart: true,
				ClearErrors:   true,
				KeepCached:    true,
			},
			sIoeOpts: &proxy.IoErrOpts{
				ResetOnNew: true,
				ListRate:   7, StatRate: 7, GetReadCloserRate: 7, ReadRate: 70, RCloseRate: 7,
			},
			tIoeOpts: &proxy.IoErrOpts{
				ResetOnNew: true,
				ListRate:   7, StatRate: 7, GetReadCloserRate: 7, ReadRate: 70, RCloseRate: 7,
				MkdirRate: 7, SetStatRate: 7, GetWriteCloserRate: 7, WriteRate: 70, WCloseRate: 7, RmRate: 7, SymlinkRate: 7,
			},
			restartCb: func(phase string, count int) bool {
				if phase != "phase1" || count >= 2 {
					return false
				}
				return true
			},
		},
		{
			label: "update once - inventory - no keep cache then again",
			ftgen: ftGenTiny,
			owo: &config.OpeLogOptionsType{
				SyncOptionsType: config.SyncOptionsType{Check: true},
				InvCsAlgos:      "md5",
			},
			hasLoadInv: true,
			custoTest:  custTestNoKeepCached,
		},
		{
			label: "update once - new session",
			ftgen: ftGenAugSmall,
			conc:  4,
			owo: &config.OpeLogOptionsType{
				SyncOptionsType: config.SyncOptionsType{Check: true},
			},
			session2:  "ses2",
			custoTest: custTestNewSession,
		},
	}

	skipDefault := false
	for it, owt := range owts {
		if skipDefault {
			if !owt.unSkipped {
				owt.lgr.Info("skipped", "test", owt.label)
				continue
			}
		}
		owt.label = fmt.Sprintf("%d - %s", it+1, owt.label)
		owt.t = t
		err := owt.runOwTest()
		require.NoError(t, err, err)
	}
}

// =======================================================================
// =======================================================================
// =======================================================================
// =======================================================================
// =======================================================================
func TestSingleOplWalkers(t *testing.T) {
	//t.Skip("for setup only")
	cliLgr, _ := common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	_ = cliLgr
	owts := []owTest{
		{
			label: "test - update once - new session",
			lgr: common.DbgLogger(),
			ftgen: ftGenAugSmall,
			owo: &config.OpeLogOptionsType{
				SyncOptionsType: config.SyncOptionsType{Check: true},
			},
			session2:        "ses2",
			custoTest:       custTestNewSession,
			specialInterest: []string{""},
			speIntPhase:     "phase3nono",
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
