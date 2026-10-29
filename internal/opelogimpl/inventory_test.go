package opelogimpl

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestConvs(t *testing.T) {
	tm := time.Now()
	sTm, err := dcTime(expDispMtime(tm))
	require.NoError(t, err)
	require.Equal(t, expDispMtime(tm), expDispMtime(time.Unix(sTm, 0)))
}

func TestInventoryMediumCsvImport(t *testing.T) {
	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 100, 3000, 2, 4096, false, 2))
	ctd := t.TempDir()
	cPath := path.Join(ctd, "invDump.csv")
	require.NoError(t, InventoryCsvExport(std, cPath, "md5,sha256,sha512"))
	ttd := t.TempDir()
	oplm, err := makeTestM2fManager(common.GetLogger(), path.Join(ctd, "m2f.opl"), 0, true, std, ttd)
	require.NoError(t, err)
	require.Contains(t, oplm.GetOlm().Sessions, "di")
	iTs := oplm.GetOlm().Sessions["di"].ReferenceTime

	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, ""))
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5,sha512"))
	err = InventoryCsvImport(oplm, iTs, cPath, "md5,sha3_256,sha512")
	require.Error(t, err)
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5,sha512"))
	require.NoError(t, oplm.Close())

	require.Contains(t, oplm.GetOlm().Sessions, "di")
	iTs = oplm.GetOlm().Sessions["di"].ReferenceTime
	require.Contains(t, oplm.GetOlm().Sessions, "ds")
	require.Contains(t, oplm.GetOlm().Sessions, "di")
}

func TestInventoryAugSmallCsvImport(t *testing.T) {
	std := t.TempDir()
	_, _, err := common.MakeAugmentedTestFilesTree(std, 7, 100, 16, 1024)
	require.NoError(t, err)
	ctd := t.TempDir()
	cPath := path.Join(ctd, "invDump.csv")
	require.NoError(t, InventoryCsvExport(std, cPath, "md5,sha256,sha512"))

	ttd := t.TempDir()
	oplm, err := makeTestM2fManager(common.GetLogger(), path.Join(ctd, "m2f.opl"), 0, true, std, ttd)
	require.NoError(t, err)
	require.Contains(t, oplm.GetOlm().Sessions, "di")
	iTs := oplm.GetOlm().Sessions["di"].ReferenceTime

	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, ""))
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5,sha512"))
	err = InventoryCsvImport(oplm, iTs, cPath, "md5,sha3_256,sha512")
	require.Error(t, err)
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5,sha512"))
	require.NoError(t, oplm.Close())

	oplm2, err := openTestM2fManagerRO(common.GetLogger(), path.Join(ctd, "m2f.opl"), true)
	require.NoError(t, err)
	require.Contains(t, oplm2.GetOlm().Sessions, "ds")
	require.Contains(t, oplm2.GetOlm().Sessions, "di")
}
