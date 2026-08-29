package opelogimpl

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type largeQ struct {
	mx         sync.Mutex
	dq chan any
	dir        string
	segSize    int
	curOffset  int
	lastOffset int
	closed     bool
	curEntries []string
	lastEntries []string
}

// Close implements [opelog.Queue].
func (lq *largeQ) Close() error {
	panic("unimplemented")
}

// Dequeue implements [opelog.Queue].
func (lq *largeQ) Dequeue() (string, error) {
	lq.mx.Lock()
	for lq.curOffset == lq.lastOffset {
		lq.mx.Unlock()
		<- lq.dq
		lq.mx.Lock()
	}
	defer lq.mx.Unlock()
}

// Enqueue implements [opelog.Queue].
func (lq *largeQ) Enqueue(s string) error {
	if strings.Contains(s, "\n") {
		return errors.New("largeQ.Enqueue optimization forbids \\n character")
	}
	lq.mx.Lock()
	defer lq.mx.Unlock()
	lastEntries := lq.lastEntries
	curSeg := lq.curOffset / lq.segSize
	lastSeg := lq.lastOffset / lq.segSize
	if lastSeg == curSeg {
		lastEntries = lq.curEntries
	}
	segOffset := (len(lastEntries) + 1) % lq.segSize
	if segOffset == 0 {
		if lastSeg != curSeg {
			fp := path.Join(lq.dir, fmt.Sprintf(".largeQ-%d.txt", segOffset))
			if err := common.WriteFile(fp, []byte(strings.Join(lastEntries, "\n"))); err != nil {
				return fmt.Errorf("largeQ.Enqueue: write %s error %v", fp, err)
			}
		}
		lq.lastEntries = make([]string, lq.segSize)
		lastEntries = lq.lastEntries
	}
	lastEntries[segOffset] = s
	if lq.curOffset == lq.lastOffset {
		lq.dq <- ""
	}
	lq.lastOffset += 1
	return nil
}

func MakeLargeQ(dir string, segSize int) (opelog.Queue, error) {
	dq := make(chan any, 1)
	if segSize == 0 {
		segSize = 1000000
	}
	curEntries := make([]string, segSize)
	return &largeQ{dq: dq, dir: dir, segSize: segSize, curEntries: curEntries}, nil
}
