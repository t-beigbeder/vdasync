package walker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/grpcclient"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/otvl_dtacsy/internal/remote"
)

func TestBasicWalker(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	td2 := t.TempDir()
	cli, cFunc, err := remote.GrpcGetTestClient()
	require.Nil(t, err)
	defer cFunc()
	dssa1 := localfiles.MakeLocalFilesDssa()
	dssa2 := grpcclient.MakeGrpcClient(context.Background(), cli)
	_, err = dssa1.List(common.OsPath2DssPath(td1))
	require.Nil(t, err)
	_, err = dssa2.List(common.OsPath2DssPath(td2))
	require.Nil(t, err)
	walker := MakeWalker(lgr, 5, dssa1, common.OsPath2DssPath(td1), "TestBasicWalker", td1, dssa2, td2)
	walker.Run()
}