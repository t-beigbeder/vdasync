package opelogimpl

import (
	"errors"
	"fmt"
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
	for lq.curOffset == lq.lastOffset {
		lq.mx.Unlock()
		if lq.closed {
			return "", errors.New("largeQ.Dequeue: all is read on closed queue")
		}
		lq.dq <- true
		lq.mx.Lock()
	}
	defer lq.mx.Unlock()
	s := lq.curEntries[lq.curOffset]
	lq.curOffset += 1
	if lq.curOffset%lq.segSize != 0 {
		return s, nil
	}
	curSeg := lq.curOffset / lq.segSize
	lastSeg := lq.lastOffset / lq.segSize
	if lq.curOffset%lq.segSize != 0 || curSeg == lastSeg {
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
	if segOffset == 0 {
		fp := path.Join(lq.dir, fmt.Sprintf(".largeQ-%d.txt", segOffset))
		if err := common.WriteFile(fp, []byte(strings.Join(lastEntries, "\n"))); err != nil {
			return fmt.Errorf("largeQ.Enqueue: write %s error %v", fp, err)
		}
		lq.lastEntries = make([]string, lq.segSize)
		lastEntries = lq.lastEntries
	}
	lastEntries[segOffset] = s
	lq.lastOffset += 1
	if lq.curOffset+1 == lq.lastOffset {
		select {
		case <-lq.dq:
		default:
		}
	}
	return nil
}

func MakeLargeQ(dir string, segSize int) (opelog.Queue, error) {
	dq := make(chan bool)
	if segSize == 0 {
		segSize = 1000000
	}
	curEntries := make([]string, segSize)
	return &largeQ{dq: dq, dir: dir, segSize: segSize, curEntries: curEntries}, nil
}
