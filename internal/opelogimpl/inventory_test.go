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

func TestInventoryCsvImport(t *testing.T) {
	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 100, 3000, 2, 4096, false, 2))
	ctd := t.TempDir()
	cPath := path.Join(ctd, "invDump.csv")
	require.NoError(t, InventoryCsvExport(std, cPath, "md5,sha256,sha512"))
	oplm, err := MakeM2fManager(path.Join(ctd, "m2f.opl"))
	require.NoError(t, err)
	oplmi, ok := oplm.(*m2fOplMng)
	require.True(t, ok)
	oplmi.testMarsh = true
	ttd := t.TempDir()
	_, iTs, err := oplm.Create("ds", "di", std, ttd)
	require.NoError(t, err)
	_, _, err = oplm.Open("ds", "di", false)
	require.NoError(t, err)
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, ""))
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5,sha512"))
	err = InventoryCsvImport(oplm, iTs, cPath, "md5,sha3_256,sha512")
	require.Error(t, err)
	require.NoError(t, InventoryCsvImport(oplm, iTs, cPath, "md5,sha512"))
	require.NoError(t, oplm.Close())
	oplm2, err := MakeM2fManager(path.Join(ctd, "m2f.opl"))
	require.NoError(t, err)
	_, _, err = oplm2.Open("ds", "di", true)
	require.NoError(t, err)
}
