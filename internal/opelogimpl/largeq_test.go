package opelogimpl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLqSimple(t *testing.T) {
	td := t.TempDir()
	lq, err := MakeLargeQ(td)
	require.NoError(t, err)
	for i := range 1000002 {
		require.NoError(t, lq.Enqueue(fmt.Sprintf("%7d", i)))
	}
	for i := range 1000002 {
		s, err := lq.Dequeue()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%7d", i), s)
	}

	require.NoError(t, lq.Close())
}