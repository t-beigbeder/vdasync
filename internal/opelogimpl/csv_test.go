package opelogimpl

import (
	"log/slog"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

func TestCsvExport(t *testing.T) {
	var (
		lgr *slog.Logger
		err error
	)
	lgr = common.GetLogger()
	// lgr = common.InfoLogger()
	// lgr, err = common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	// lgr = common.DbgLogger()
	require.NoError(t, err)

	lgr.Debug("TestCsvExport: started")
	std := t.TempDir()
	ltd := t.TempDir()
	ttd := t.TempDir()

	// require.NoError(t, ftGenTiny(std))
	// require.NoError(t, ftGenMedium(std))
	require.NoError(t, ftGenAugSmall(std))
	lgr.Debug("TestCsvExport: FileTreeGenerated")

	cPath := path.Join(ltd, "invDump.csv")
	require.NoError(t, InventoryCsvExport(std, cPath, "md5"))
	oplm, err := makeTestM2fManager(path.Join(ltd, "m2f.opl"), 5, true, std, ttd)
	require.NoError(t, err)
	require.Contains(t, oplm.Olm().Sessions, "di")
	iTs := oplm.Olm().Sessions["di"].ReferenceTime
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5"))
	require.NoError(t, oplm.Close())

	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&config.OpeLogOptionsType{
			SyncOptionsType: config.SyncOptionsType{
				Check: true,
			},
			Goals:        "update",
			ResetTimeout: -1,
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

	csvPath := path.Join(ltd, "oplm.csv")
	require.NoError(t, OplCsvExport(lgr, oplmRo, csvPath, RPT_SYNTHETIC, "di", "ds"))
}
