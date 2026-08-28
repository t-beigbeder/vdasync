package opelogimpl

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestM2fOpeLogs(t *testing.T) {
	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 250, 25000, 2, 32768, false, 0))

	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := MakeM2fManager(path.Join(ltd, "m2f.opl"), std, ttd)
	require.NoError(t, err)
	require.NoError(t, olm.Init(std, ttd))
	require.NoError(t, olm.NewSession())
}
