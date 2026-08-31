package opelogimpl

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestLqSimple(t *testing.T) {
	lgr := common.DbgLogger()
	td := t.TempDir()
	lq, err := MakeLargeQ(lgr, td, 10000)
	require.NoError(t, err)
	lgr.Debug("start")
	for i := range 100000 {
		require.NoError(t, lq.Enqueue(fmt.Sprintf("%7d", i)))
	}
	lq.Enqueue("EOF")
	lgr.Debug("enqueued")

	for i := range 100000 {
		s, err := lq.Dequeue()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%7d", i), s)
	}
	s, err := lq.Dequeue()
	require.NoError(t, err)
	require.Equal(t, "EOF", s)

	require.NoError(t, lq.Close())
	lgr.Debug("end")
}

func TestLqConcur(t *testing.T) {
	lgr := common.DbgLogger()
	td := t.TempDir()
	i := 3
	_ = i
	lq, err := MakeLargeQ(lgr, td, 10000)
	require.NoError(t, err)
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
					if !strings.HasPrefix(err.Error(), "largeQ.Dequeue: all is read on closed queue") {
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

func TestLqBackAndForth(t *testing.T) {
	lgr := common.DbgLogger()
	td := t.TempDir()
	lq, err := MakeLargeQ(lgr, td, 10000)
	require.NoError(t, err)
	lgr.Debug("start")
	for i := range 30000 {
		require.NoError(t, lq.Enqueue(fmt.Sprintf("%7d", i)))
	}
	lgr.Debug("enqueued 30000")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		lgr.Debug("go subroutine started")
		for i := range 100000 {
			s, err := lq.Dequeue()
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("%7d", i), s)
			if i > 0 && i%10000 == 0 {
				lgr.Debug("dequeued", "i", i)
			}
		}
		s, err := lq.Dequeue()
		require.NoError(t, err)
		require.Equal(t, "EOF", s)
		lgr.Debug("go subroutine ended")
		wg.Done()
	}()

	time.Sleep(400 * time.Millisecond)
	for i := range 30000 {
		require.NoError(t, lq.Enqueue(fmt.Sprintf("%7d", i+30000)))
	}
	lgr.Debug("enqueued 60000")

	time.Sleep(400 * time.Millisecond)
	for i := range 40000 {
		require.NoError(t, lq.Enqueue(fmt.Sprintf("%7d", i+60000)))
	}
	lgr.Debug("enqueued 100000")
	lq.Enqueue("EOF")

	require.NoError(t, lq.Close())
	lgr.Debug("closed")

	wg.Wait()
	lgr.Debug("done")
}
