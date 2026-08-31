package opelogimpl

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type largeQ struct {
	mx          sync.Mutex
	dq          chan bool
	lgr         *slog.Logger
	dir         string
	segSize     int
	curOffset   int
	lastOffset  int
	closed      bool
	curEntries  []string
	lastEntries []string
}

// Close implements [opelog.Queue].
func (lq *largeQ) Close() error {
	lq.mx.Lock()
	defer lq.mx.Unlock()
	if lq.closed {
		return errors.New("largeQ.Close: already done")
	}
	lq.closed = true
	if lq.curOffset == lq.lastOffset {
		select {
		case <-lq.dq:
		default:
		}
	}
	return nil
}

// Dequeue implements [opelog.Queue].
func (lq *largeQ) Dequeue() (string, error) {
	lq.mx.Lock()
	wasBlocked := false
	for lq.curOffset == lq.lastOffset {
		lq.mx.Unlock()
		if lq.closed {
			return "", errors.New("largeQ.Dequeue: all is read on closed queue")
		}
		lq.lgr.Debug("largeQ.Dequeue", "curOffset", lq.curOffset, "lastOffset", lq.lastOffset, "lq.dq <-", true)
		lq.dq <- true
		lq.lgr.Debug("largeQ.Dequeue", "curOffset", lq.curOffset, "lastOffset", lq.lastOffset, "lq.dq done", true)
		lq.mx.Lock()
		lq.lgr.Debug("largeQ.Dequeue", "curOffset", lq.curOffset, "lastOffset", lq.lastOffset, "lq.dq done", "locked")
		wasBlocked = true
	}
	if wasBlocked {
		lq.lgr.Debug("largeQ.Dequeue", "curOffset", lq.curOffset, "lastOffset", lq.lastOffset, "lq.dq done", "wasBlocked")
	}
	defer lq.mx.Unlock()
	segOffset := lq.curOffset % lq.segSize
	curSeg := lq.curOffset / lq.segSize
	lastSeg := lq.lastOffset / lq.segSize
	if segOffset != 0 || curSeg == 0 {
		s := lq.curEntries[segOffset]
		lq.curOffset += 1
		return s, nil
	}
	if curSeg == lastSeg {
		copy(lq.curEntries, lq.lastEntries)
		s := lq.curEntries[segOffset]
		lq.curOffset += 1
		return s, nil
	}
	fp := path.Join(lq.dir, fmt.Sprintf(".largeQ-%d.txt", curSeg))
	bs, err := common.UnsafeLoadFile(fp)
	if err != nil {
		return "", fmt.Errorf("largeQ.Dequeue: read %s error %v", fp, err)
	}
	if err := os.Remove(fp); err != nil {
		return "", fmt.Errorf("largeQ.Dequeue: remove %s error %v", fp, err)
	}
	lns := strings.Split(string(bs), "\n")
	if len(lns) > lq.segSize {
		return "", fmt.Errorf("largeQ.Dequeue: read %s len %d", fp, len(lns))
	}
	copy(lq.curEntries, lns)
	s := lq.curEntries[segOffset]
	lq.curOffset += 1
	return s, nil
}

// Enqueue implements [opelog.Queue].
func (lq *largeQ) Enqueue(s string) error {
	if strings.Contains(s, "\n") {
		return errors.New("largeQ.Enqueue optimization forbids \\n character")
	}
	lq.mx.Lock()
	defer lq.mx.Unlock()
	if lq.closed {
		return errors.New("largeQ.Enqueue: write on closed queue")
	}
	lastEntries := lq.lastEntries
	curSeg := lq.curOffset / lq.segSize
	lastSeg := lq.lastOffset / lq.segSize
	if lastSeg == curSeg {
		lastEntries = lq.curEntries
	}
	segOffset := lq.lastOffset % lq.segSize
	lastEntries[segOffset] = s
	if segOffset == lq.segSize-1 {
		if lastSeg != 0 {
			fp := path.Join(lq.dir, fmt.Sprintf(".largeQ-%d.txt", lastSeg))
			if err := common.WriteFile(fp, []byte(strings.Join(lastEntries, "\n"))); err != nil {
				return fmt.Errorf("largeQ.Enqueue: write %s error %v", fp, err)
			}
		}
		lq.lastEntries = make([]string, lq.segSize)
	}
	lq.lastOffset += 1
	if lq.curOffset+1 == lq.lastOffset {
		lq.lgr.Debug("largeQ.Enqueue")
		lq.lgr.Debug("largeQ.Enqueue", "curOffset", lq.curOffset, "lastOffset", lq.lastOffset, "<-lq.dq", "?")
		select {
		case vt := <-lq.dq:
			lq.lgr.Debug("largeQ.Enqueue", "curOffset", lq.curOffset, "lastOffset", lq.lastOffset, "<-lq.dq", vt)
		default:
			lq.lgr.Debug("largeQ.Enqueue", "curOffset", lq.curOffset, "lastOffset", lq.lastOffset, "<-lq.dq", "default")
		}
	}
	return nil
}

func MakeLargeQ(lgr *slog.Logger, dir string, segSize int) (opelog.Queue, error) {
	dq := make(chan bool)
	if segSize == 0 {
		segSize = 1000000
	}
	curEntries := make([]string, segSize)
	return &largeQ{dq: dq, lgr: lgr, dir: dir, segSize: segSize, curEntries: curEntries}, nil
}
