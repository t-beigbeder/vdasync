package remote

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/opegrpc"
)

func TestRunOpeDssaServer(t *testing.T) {
	td := t.TempDir()
	t.Chdir(td)
	common.WriteFile(t.Name()+".txt", []byte(t.Name()+"\n"))
	port, cFunc, err := RunOpeDssaServer(common.GetLogger(), context.Background(), testHost, 0, nil, localfiles.MakeLocalFilesDssa(), nil)
	require.Nil(t, err)
	cli, conn, err := NewOpeDssaClient(fmt.Sprintf("%s:%d", testHost, port), nil)
	require.Nil(t, err)
	defer conn.Close()

	rr, err := cli.Ready(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.True(t, rr.Value)
	rv, err := cli.Version(context.Background(), &opegrpc.Empty{})
	require.Nil(t, err)
	require.Equal(t, "0.1", rv.Value)

	wd, err := os.Getwd()
	require.Nil(t, err)
	rl, err := cli.List(context.Background(), &dssagrpc.Path{Path: wd})
	require.Nil(t, err)
	require.Equal(t, 1, len(rl.Entries))
	require.False(t, rl.Entries[0].IsDir)
	_, name := path.Split(rl.Entries[0].Path)
	require.Equal(t, t.Name()+".txt", name)

	cFunc()
	time.Sleep(50 * time.Millisecond)
}
