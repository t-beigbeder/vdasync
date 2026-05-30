package sftpc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
)

func TestSftpStuff(t *testing.T) {
	ds := GetSftpDss(t)
	fis, err := ds.List("/")
	require.NoError(t, err)
	require.Zero(t, len(fis))
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d1", IsDir: true}))
}

func TestBasicDirsAndFiles(t *testing.T) {
	SkipIf(t)
}