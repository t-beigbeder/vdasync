package walker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/grpcclient"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/remote"
)

func TestBasicWalker(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	require.Nil(t, err)
	lgr.Debug("TestBasicWalker", "td1", td1, "sad", sad, "saf", saf)

	dssa1 := localfiles.MakeLocalFilesDssa()
	// _, err = dssa1.List(common.OsPath2DssPath(td1))
	// require.Nil(t, err)
	td2 := t.TempDir()
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.Nil(t, err)
	defer cFunc()
	dssa2 := grpcclient.MakeGrpcClient(lgr, context.Background(), cli)
	// _, err = dssa2.List(common.OsPath2DssPath(td2))
	// require.Nil(t, err)

	startDe := func(pe *ProcessedEntry, noLstatOnList bool) []*dssa.DataEntry {
		des, _ := pe.wi.ds.List(pe.DataEntry.Path)
		return des
	}
	startNde := func(pe *ProcessedEntry) {
	}
	walker := MakeWalker(lgr, 4, dssa1, startDe, startNde, nil, nil, nil, nil, "TestBasicWalker", td1, dssa2, td2)
	walker.Run(&dssa.DataEntry{Path: td1, IsDir: true})
	lgr.Debug("TestBasicWalker: done")
}
