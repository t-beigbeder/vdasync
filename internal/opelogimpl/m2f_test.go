package opelogimpl

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func TestM2fOpeLogs(t *testing.T) {
	lgr := common.DbgLogger()
	lgr.Debug("TestM2fOpeLogs: start")

	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 250, 25000, 2, 32768, false, 0))
	lgr.Debug("TestM2fOpeLogs: FileTreeGenerated")

	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := MakeM2fManager(path.Join(ltd, "m2f.opl"), std, ttd)
	require.NoError(t, err)
	require.NoError(t, olm.Init(std, ttd))
	require.NoError(t, olm.NewSession())
	olm.PutLogicalEntry("", &opelog.LogicalEntry{})
	require.NoError(t, olm.EndSession())
	lgr.Debug("TestM2fOpeLogs: end")
}
