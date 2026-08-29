package opelogimpl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestLqSimple(t *testing.T) {
	td := t.TempDir()
	lq, err := MakeLargeQ(td, 10000)
	require.NoError(t, err)
	lgr := common.DbgLogger()
	lgr.Debug("start")
	for i := range 20000 {
		require.NoError(t, lq.Enqueue(fmt.Sprintf("%7d", i)))
	}
	lq.Enqueue("EOF")
	for i := range 20000 {
		s, err := lq.Dequeue()
		if err != nil {
			lgr.Debug("this")
		}
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%7d", i), s)
	}
	s, err := lq.Dequeue()
	require.NoError(t, err)
	require.Equal(t, "EOF", s)

	require.NoError(t, lq.Close())
	lgr.Debug("end")
}
