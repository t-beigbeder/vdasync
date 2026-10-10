package common

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLimBck(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	td := t.TempDir()
	path_ := path.Join(td, "TestLimBck.dat")
	require.NoError(t, MakeTestFile(path_, 1024))
	require.NoError(t, LimitedBackup(path_, 0))
	time.Sleep(time.Second)
	require.NoError(t, MakeTestFile(path_, 1025))
	require.NoError(t, LimitedBackup(path_, 0))
	time.Sleep(time.Second)
	require.NoError(t, MakeTestFile(path_, 1026))
	require.NoError(t, LimitedBackup(path_, 0))
	pathFake := path.Join(td, "TestLimBck.dat.20060102-150405")
	require.NoError(t, MakeTestFile(pathFake, 1023))
	require.NoError(t, LimitedBackup(path_, 2))
	time.Sleep(time.Second)
	require.NoError(t, MakeTestFile(path_, 1027))
	require.NoError(t, LimitedBackup(path_, 2))
}
