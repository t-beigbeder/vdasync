package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExclIncl(t *testing.T) {
	inclRegs, err := ReFromSlice([]string{}, "incl")
	exclRegs, err := ReFromSlice([]string{}, "excl")
	require.NoError(t, err)
	require.True(t, IsIncluded("", inclRegs, exclRegs))
	inclRegs, err = ReFromSlice([]string{"^d00/"}, "incl")
	require.True(t, IsIncluded("d00/f1.txt", inclRegs, exclRegs))
	exclRegs, err = ReFromSlice([]string{"^d00/f1"}, "incl")
	require.False(t, IsIncluded("d00/f1.txt", inclRegs, exclRegs))
	inclRegs, err = ReFromSlice([]string{"^d00/", "^d01/"}, "incl")
	require.True(t, IsIncluded("d01/f1.txt", inclRegs, exclRegs))
	inclRegs, err = ReFromSlice([]string{}, "incl")
	require.True(t, IsIncluded("d01/f1.txt", inclRegs, exclRegs))
	require.False(t, IsIncluded("d00/f1.txt", inclRegs, exclRegs))
}
