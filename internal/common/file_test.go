package common

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
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

func TestAccessRights(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestAccessRights.dat")
	require.Nil(t, WriteFile(ft, []byte(t.Name())))
	fi, ugIds, ugoRights, err := GetFileStat(ft)
	require.Nil(t, err)
	require.Equal(t, len(t.Name()), int(fi.Size()))
	ugoRights[1] = dssa.Rights{Read: true}
	ugoRights[2] = dssa.Rights{}
	err = SetAccessRights(ft, ugIds, ugoRights)
	require.Nil(t, err)
	fi, ugIds, ugoRights, err = GetFileStat(ft)
	require.Nil(t, err)
	require.False(t, ugoRights[1].Write)
}