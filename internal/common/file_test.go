package common

import (
	"github.com/stretchr/testify/require"
	"path"
	"testing"
)

func TestFileFunctions(t *testing.T) {
	var (
		sz  int64
		err error
		bs  []byte
	)
	ft := path.Join(t.TempDir(), "TestFileFunctions.dat")
	require.False(t, FileExists(ft))
	require.Nil(t, WriteFile(ft, []byte(t.Name())))
	sz, err = FileSize(ft)
	require.Nil(t, err)
	require.Equal(t, len(t.Name()), int(sz))
	bs, err = LoadFile(ft)
	require.Nil(t, err)
	require.Equal(t, []byte(t.Name()), bs)
	var bs2 = [MaxLoadFileSize + 1]byte{}
	bs = bs2[:]
	require.Nil(t, WriteFile(ft, bs))
	sz, err = FileSize(ft)
	require.Nil(t, err)
	require.Equal(t, MaxLoadFileSize+1, int(sz))
	bs, err = LoadFile(ft)
	require.NotNil(t, err)
}
