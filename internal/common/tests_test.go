package common

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	GetLogger().Info("a message", "with", "that")
	GetLogger().Debug("another message", "that is for", "debug")
}

func TestMakeTestFile(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestFileGetPut.dat")
	for _, size := range []int64{1023, 32*1024 - 1, 32*1024*1024 - 1, 32 * 1024 * 1024} {
		err := MakeTestFile(ft, int(size))
		require.Nil(t, err)
		fi, err := os.Stat(ft)
		require.Nil(t, err)
		require.Equal(t, size, fi.Size())

	}
}
