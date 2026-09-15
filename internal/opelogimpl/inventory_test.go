package opelogimpl

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestInventoryCsvImport(t *testing.T) {
	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 100, 3000, 2, 4096, false, 2))
	ctd := t.TempDir()
	cPath := path.Join(ctd, "invDump.csv")
	require.NoError(t, InventoryCsvExport(std, cPath, "md5,sha256,sha512"))
	oplm, err := MakeM2fManager(path.Join(ctd, "m2f.opl"))
	require.NoError(t, err)
	ttd := t.TempDir()
	require.NoError(t, oplm.Create(std, ttd))
	require.NoError(t, oplm.Open(false))
	require.NoError(t, InventoryCsvImport(oplm, cPath, ""))
	require.NoError(t, InventoryCsvImport(oplm, cPath, "md5,sha512"))
	require.Error(t, InventoryCsvImport(oplm, cPath, "md5,sha3_256,sha512"))
}
