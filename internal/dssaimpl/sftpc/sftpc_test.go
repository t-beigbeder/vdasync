package sftpc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
)

func TestSftpStuff(t *testing.T) {
	ds := GetSftpDss(t)
	des, err := ds.List("/")
	require.NoError(t, err)
	require.Zero(t, len(des))
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d1", IsDir: true}))
	des, err = ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 1, len(des))
}

func TestBasicDirsAndFiles(t *testing.T) {
	SkipIf(t)
}
