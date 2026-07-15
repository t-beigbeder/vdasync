package cli

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

func TestServiceList(t *testing.T) {
	dss := localfiles.MakeLocalFilesDssa()
	td := t.TempDir()
	lgr := common.DbgLogger()
	_, _, err := common.AugmentTestFilesTree(td)
	require.NoError(t, err)
	des, err := dss.List(path.Join(td))
	require.NoError(t, err)
	for _, de := range des {
		lgr.Debug("TestDeList", "de", common.DataEntryList(de, false, ""))
	}
	require.NoError(t,
		DoService(&ServiceCtx{
			Cmd:         "list",
			Dss:         dss,
			Root:        td,
			IsRecur:     false,
			IsCheck:     true,
			Concurrency: 0,
			Lgr:         lgr,
			OutFile:     os.Stdout,
		}),
	)
	require.NoError(t,
		DoService(&ServiceCtx{
			Cmd:         "list",
			Dss:         dss,
			Root:        td,
			IsRecur:     true,
			IsCheck:     true,
			Concurrency: 0,
			Lgr:         lgr,
			OutFile:     os.Stdout,
		}),
	)
}
