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
	lgr         *slog.Logger
	dir         string
	segSize     int
	mx          sync.Mutex
	dq          chan bool
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
	segOffset := lq.lastOffset % lq.segSize
	if segOffset > 0 {
		lastSeg := lq.lastOffset / lq.segSize
		fp := path.Join(lq.dir, fmt.Sprintf(".largeQ-%d.txt", lastSeg))
		if err := common.WriteFile(fp, []byte(strings.Join(lq.lastEntries[0:segOffset], "\n"))); err != nil {
			return fmt.Errorf("largeQ.Enqueue: write %s error %v", fp, err)
		}
	}
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
	for lq.curOffset == lq.lastOffset {
		lq.mx.Unlock()
		if lq.closed {
			return "", errors.New("largeQ.Dequeue: all is read on closed queue")
		}
		lq.dq <- true
		lq.mx.Lock()
	}
	defer lq.mx.Unlock()

	segOffset := lq.curOffset % lq.segSize
	if segOffset != 0 {
		s := lq.curEntries[segOffset]
		lq.curOffset += 1
		return s, nil
	}

	curSeg := lq.curOffset / lq.segSize
	lastSeg := lq.lastOffset / lq.segSize
	if curSeg > 0 {
		fp := path.Join(lq.dir, fmt.Sprintf(".largeQ-%d.txt", curSeg-1))
		os.Remove(fp)
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
	curSeg := lq.curOffset / lq.segSize
	lastSeg := lq.lastOffset / lq.segSize
	segOffset := lq.lastOffset % lq.segSize
	lq.lastEntries[segOffset] = s
	if curSeg == lastSeg {
		lq.curEntries[segOffset] = s
	}
	if segOffset == lq.segSize-1 {
		fp := path.Join(lq.dir, fmt.Sprintf(".largeQ-%d.txt", lastSeg))
		if err := common.WriteFile(fp, []byte(strings.Join(lq.lastEntries, "\n"))); err != nil {
			return fmt.Errorf("largeQ.Enqueue: write %s error %v", fp, err)
		}
		lq.lastEntries = make([]string, lq.segSize)
	}
	lq.lastOffset += 1
	if lq.curOffset+1 == lq.lastOffset {
		select {
		case <-lq.dq:
		default:
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
	lastEntries := make([]string, segSize)
	return &largeQ{
		dq: dq, lgr: lgr, dir: dir, segSize: segSize,
		curEntries: curEntries, lastEntries: lastEntries,
	}, nil
}
