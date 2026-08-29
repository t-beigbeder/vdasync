package opelogimpl

import (
	"fmt"
	"strings"
	"sync"
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
	for i := range 100000 {
		require.NoError(t, lq.Enqueue(fmt.Sprintf("%7d", i)))
	}
	lq.Enqueue("EOF")
	lgr.Debug("enqueued")
	for i := range 100000 {
		s, err := lq.Dequeue()
		if err != nil {
			lgr.Debug("this")
		}
		require.NoError(t, err)
		if fmt.Sprintf("%7d", i) != s {
			lgr.Debug("here")
		}
		require.Equal(t, fmt.Sprintf("%7d", i), s)
	}
	s, err := lq.Dequeue()
	require.NoError(t, err)
	require.Equal(t, "EOF", s)

	require.NoError(t, lq.Close())
	lgr.Debug("end")
}

func TestLqConcur(t *testing.T) {
	td := t.TempDir()
	lq, err := MakeDQ(td, 10000)
	require.NoError(t, err)
	lgr := common.DbgLogger()
	lgr.Debug("start")
	for i := range 100000 {
		require.NoError(t, lq.Enqueue(fmt.Sprintf("%7d", i)))
	}
	lq.Enqueue("EOF")
	lgr.Debug("enqueued")
	var wg sync.WaitGroup
	var counts [10]int
	done := make(chan any, 1)
	for i := range 10 {
		wg.Add(1)
		go func(j int) {
			for {
				s, err := lq.Dequeue()
				if err != nil {
					if !strings.HasPrefix(err.Error(), "queue is closed") {
						lgr.Error("concur", "i", j, "err", err)
					}
					break
				}
				if s == "EOF" {
					close(done)
					break
				}
				counts[j] += 1
			}
			wg.Done()
		}(i)
	}
	<-done
	lq.Close()
	wg.Wait()
	count := 0
	for i := range 10 {
		count += counts[i]
	}
	lgr.Debug("end", "count", count)
}
