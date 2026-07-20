package cli

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/grpcclient"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/remote"
)

func TestServiceList(t *testing.T) {
	dss := localfiles.MakeLocalFilesDssa()
	td := t.TempDir()
	lgr := common.GetLogger()
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
			OutFile:     common.GetTestOut(),
		}),
	)
	require.NoError(t,
		DoService(&ServiceCtx{
			Cmd:         "list",
			Dss:         dss,
			Root:        td,
			IsRecur:     true,
			IsCheck:     true,
			Latency:     "1us",
			Concurrency: 0,
			Lgr:         lgr,
			OutFile:     common.GetTestOut(),
		}),
	)
}

func TestServiceLatencyRaw(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	// will take ~ 5s
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.Nil(t, err)
	defer cFunc()
	dgc := grpcclient.MakeGrpcClient(common.GetLogger(), context.Background(), cli)
	require.NoError(t,
		DoService(&ServiceCtx{
			Cmd:         "latency",
			Dss:         dgc,
			Latency:     "1s",
			Count:       5,
			Concurrency: 0,
		}),
	)
}

func TestServiceLatencySimulNet(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	// will take ~ 5s
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.Nil(t, err)
	defer cFunc()
	dgc := grpcclient.MakeGrpcClient(common.GetLogger(), context.Background(), cli)
	require.NoError(t,
		DoService(&ServiceCtx{
			Cmd:         "latency",
			Dss:         dgc,
			Latency:     "80ms",
			Count:       20000,
			Concurrency: 320,
		}),
	)
}

func TestServiceLatencySimulCompute(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	// will take ~ 5s
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.Nil(t, err)
	defer cFunc()
	dgc := grpcclient.MakeGrpcClient(common.GetLogger(), context.Background(), cli)
	require.NoError(t,
		DoService(&ServiceCtx{
			Cmd:         "latency",
			Dss:         dgc,
			Latency:     "6400ms",
			Count:       250,
			Concurrency: 320,
		}),
	)
}

func TestOtherServices(t *testing.T) {
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.Nil(t, err)
	defer cFunc()
	dgc := grpcclient.MakeGrpcClient(common.GetLogger(), context.Background(), cli)
	require.NoError(t,
		DoService(&ServiceCtx{
			Cmd: "version",
			Dss: dgc,
		}),
	)
	require.NoError(t,
		DoService(&ServiceCtx{
			Cmd: "shutdown",
			Dss: dgc,
		}),
	)
}
