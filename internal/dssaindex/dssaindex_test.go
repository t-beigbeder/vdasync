package dssaindex

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
)

func TestXxx(t *testing.T) {
	mid := MakeMemIndexDssa()
	des, err := mid.List("/")
	require.NoError(t, err)
	require.Zero(t, len(des))

	mid.Mkdir(&dssa.DataEntry{IsDir: true, Path: "/d1"})
	require.NoError(t, err)
	des, err = mid.List("/")
	require.NoError(t, err)
	require.Equal(t, 1, len(des))

	mid.Mkdir(&dssa.DataEntry{IsDir: true, Path: "/d1/d2"})
	require.NoError(t, err)
	des, err = mid.List("/d1")
	require.NoError(t, err)
	require.Equal(t, 1, len(des))

	wc, err := mid.GetWriteCloser("/d1/f1")
	require.NoError(t, err)
	wc.Close()

	wc, err = mid.GetWriteCloser("/d1/d2/f2")
	require.NoError(t, err)
	wc.Close()

	err = mid.Rm("/d1/d2")
	require.Error(t, err)
	require.NoError(t, mid.Rm("/d1/d2/f2"))
	require.NoError(t, mid.Rm("/d1/d2"))
}
