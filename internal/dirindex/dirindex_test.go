package dirindex

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

	mid.Put(&dssa.DataEntry{IsDir: true, Path: "/d1"})
	require.NoError(t, err)
	des, err = mid.List("/")
	require.NoError(t, err)
	require.Equal(t, 1, len(des))

	mid.Put(&dssa.DataEntry{IsDir: true, Path: "/d1/d2"})
	require.NoError(t, err)
	des, err = mid.List("/d1")
	require.NoError(t, err)
	require.Equal(t, 1, len(des))

	require.NoError(t, mid.Put(&dssa.DataEntry{Path: "/d1/f1"}))
	require.NoError(t, mid.Put(&dssa.DataEntry{Path: "/d1/d2/f2"}))

	require.Error(t, mid.Del("/d1/d2"))
	require.NoError(t, mid.Del("/d1/d2/f2"))
	require.NoError(t, mid.Del("/d1/d2"))

	_, err = mid.Get("/d1/d2")
	require.Error(t, err)
	_, err = mid.Get("/d1")
	require.NoError(t, err)
}
