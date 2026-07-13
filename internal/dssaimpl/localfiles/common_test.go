package localfiles

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestDeList(t *testing.T) {
	dss := MakeLocalFilesDssa()
	td := t.TempDir()
	lgr := common.DbgLogger()
	_, _, err := common.AugmentTestFilesTree(td)
	require.NoError(t, err)
	des, err := dss.List(path.Join(td))
	require.NoError(t, err)
	for _, de := range des {
		lgr.Debug("TestDeList", "de", common.DataEntryList(de))
	}
	des, err = dss.List(path.Join(td, "dLinks"))
	require.NoError(t, err)
	for _, de := range des {
		lgr.Debug("TestDeList", "de", common.DataEntryList(de))
	}
	des, err = dss.List(path.Join(td, "dAddFiles"))
	require.NoError(t, err)
	for _, de := range des {
		lgr.Debug("TestDeList", "de", common.DataEntryList(de))
	}
}
