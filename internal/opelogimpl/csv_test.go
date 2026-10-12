package opelogimpl

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

func TestCsvExport(t *testing.T) {
	var (
		err error
	)
	dbgLgr := common.DbgLogger()
	cliLgr, err := common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	require.NoError(t, err)
	infLgr := common.InfoLogger()
	defLgr := common.GetLogger()
	_, _, _, _ = dbgLgr, cliLgr, infLgr, defLgr
	defLgr = infLgr
	skipDefault := false

	owts := []owTest{
		{
			label:     "load - c4 small simple",
			unSkipped: false,
			lgr:       defLgr,
			ftgen:     ftGenSmall,
			conc:      0,
			owo: &config.OpeLogOptionsType{
				Goals:        "load",
				ResetTimeout: -1,
			},
			oplm:    nil,
			oplq:    nil,
			loadInv: false,
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
			owt.oplm, err = makeTestM2fManager(path.Join(owt.ltd, "m2f.opl"), 0, true, owt.std, owt.ttd)
		}
		require.NoError(t, err)
		require.NoError(t, owt.ftgen(owt.std))
		require.NoError(t, owt.invImport())
		require.NoError(t, owt.oplm.Close())
		ow := NewOplWalker(
			owt.lgr.With("test", owt.label), owt.conc,
			nil, owt.oplm, owt.owo,
			localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(),
			owt.std, owt.ttd,
			"ds", "di",
		)
		require.NoError(t, ow.Run())
		err = owt.oplm.Open(true)
		require.NoError(t, err)
		csvPath := path.Join(owt.ltd, "oplm.csv")
		require.NoError(t, OplCsvExport(owt.lgr, ow, owt.oplm, csvPath, RPT_SYNTHETIC, "ds"))
		owt.lgr.Info("exported", "csv", csvPath)
	}
}
