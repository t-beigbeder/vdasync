package opelogimpl

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestInventoryCsvExport(t *testing.T) {
	ttd := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(ttd, 100, 3000, 2, 4096, false, 2))
	ctd := t.TempDir()
	err := InventoryCsvExport(ttd, path.Join(ctd, "invDump.csv"), "md5,sha256,sha512")
	require.NoError(t, err)
}
