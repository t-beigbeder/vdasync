package common

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	sll := os.Getenv("GO_TEST_LOG_LEVEL")
	GetLogger().Error("an error message", "with", "that", "GO_TEST_LOG_LEVEL", sll)
	GetLogger().Info("a message", "with", "that")
	GetLogger().Debug("another message", "that is for", "debug")
	il := doGetLogger("INFO")
	il.Debug("not displayed")
	il.Info("displayed")
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

func TestMakeTestFilesTree(t *testing.T) {
	td := t.TempDir()
	sad, saf, err := MakeTestFilesTree(td, 7, 100, 16, 6*1024*1024)
	require.Nil(t, err)
	GetLogger().Debug("TestMakeTestFilesTree", "td", td, "sad", sad, "saf", saf)
}

func TestMakeAugmentedTestFilesTree(t *testing.T) {
	td := t.TempDir()
	sad, saf, err := MakeAugmentedTestFilesTree(td, 7, 100, 16, 6*1024*1024)
	GetLogger().Debug("TestMakeAugmentedTestFilesTree", "td", td, "sad", sad, "saf", saf)
	require.Nil(t, err)
	/*
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dAddFiles
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dAddFiles/dRemoved
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dAddFiles/dStay
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dAddFiles/fRemoved.dat
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dAddFiles/fStay.dat
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dLinks
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dLinks/dotDot.lnk
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dLinks/fRef.dat
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dLinks/fRef.lnk
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dLinks/notYet.lnk
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dMod
	/local/tmp/TestMakeAugmentedTestFilesTree4265037193/001/dau/dRO
	*/
}
