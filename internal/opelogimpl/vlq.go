package opelogimpl

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/gammazero/deque"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type veryLongQueue struct {
	lgr      *slog.Logger
	dir      string
	segSize  int
	mx       sync.Mutex
	closed   bool
	prodSubs []chan bool
	cOff     int
	pOff     int
	cEntries *deque.Deque[string]
	pEntries *deque.Deque[string]
}

func (vlq *veryLongQueue) saveEntries() error {
	fp := path.Join(vlq.dir, fmt.Sprintf(".largeQ-%d.txt", vlq.pOff/vlq.segSize-1))
	out := make([]string, 0, vlq.pEntries.Len())
	out = vlq.pEntries.AppendToSlice(out)
	if err := common.WriteFile(fp, []byte(strings.Join(out, "\n"))); err != nil {
		return fmt.Errorf("largeQ.saveEntries: write %s error %v", fp, err)
	}
	vlq.pEntries.Clear()
	return nil
}

func (vlq *veryLongQueue) prodEvent() {
	if vlq.cOff == vlq.pOff {
		for _, prodSub := range vlq.prodSubs {
			close(prodSub)
		}
		vlq.prodSubs = make([]chan bool, 0)
	}
}

// Close implements [opelog.Queue].
func (vlq *veryLongQueue) Close() error {
	vlq.mx.Lock()
	defer vlq.mx.Unlock()
	if vlq.closed {
		return errors.New("largeQ.Close: already done")
	}
	vlq.closed = true
	if vlq.pEntries.Len() > 0 {
		if err := vlq.saveEntries(); err != nil {
			return err
		}
	}
	vlq.prodEvent()
	return nil
}

// Get implements [opelog.Queue].
func (vlq *veryLongQueue) Get() (string, error) {
	vlq.mx.Lock()
	for vlq.cOff == vlq.pOff {
		if vlq.closed {
			vlq.mx.Unlock()
			return "", common.ErrReadClosedQueue
		}
		prodSub := make(chan bool)
		vlq.prodSubs = append(vlq.prodSubs, prodSub)
		vlq.mx.Unlock()
		<-prodSub
		vlq.mx.Lock()
	}
	defer vlq.mx.Unlock()

	if vlq.cOff/vlq.segSize == vlq.pOff/vlq.segSize {
		s := vlq.pEntries.At(vlq.cOff % vlq.segSize)
		vlq.cOff++
		return s, nil
	}
	if vlq.cEntries.Len() == 0 {
		fp := path.Join(vlq.dir, fmt.Sprintf(".largeQ-%d.txt", vlq.cOff/vlq.segSize))
		bs, err := common.UnsafeLoadFile(fp)
		if err != nil {
			return "", fmt.Errorf("largeQ.Get: load error %v pOff %d cOff %d", err, vlq.pOff, vlq.cOff)
		}
		if err := os.Remove(fp); err != nil {
			return "", fmt.Errorf("largeQ.Get: remove error %v", err)
		}
		lns := strings.Split(string(bs), "\n")
		if len(lns) > vlq.segSize {
			return "", fmt.Errorf("largeQ.Get: read %s len %d", fp, len(lns))
		}
		vlq.cEntries.CopyInSlice(lns)
	}
	s := vlq.cEntries.At(vlq.cOff % vlq.segSize)
	vlq.cOff++
	if vlq.cOff%vlq.segSize == 0 {
		vlq.cEntries.Clear()
	}
	return s, nil
}

// Put implements [opelog.Queue].
func (vlq *veryLongQueue) Put(s string) error {
	if strings.Contains(s, "\n") {
		return errors.New("largeQ.Put optimization forbids \\n character")
	}
	vlq.mx.Lock()
	defer vlq.mx.Unlock()
	if vlq.closed {
		return errors.New("largeQ.Put: write on closed queue")
	}
	vlq.prodEvent()
	vlq.pEntries.PushBack(s)
	vlq.pOff++
	if vlq.pEntries.Len() == vlq.segSize {
		if err := vlq.saveEntries(); err != nil {
			return err
		}
	}
	return nil
}

func MakeVeryLongQueue(lgr *slog.Logger, dir string, segSize int) (opelog.Queue, error) {
	var pEntries deque.Deque[string]
	pEntries.SetBaseCap(segSize)
	var cEntries deque.Deque[string]
	cEntries.SetBaseCap(segSize)
	return &veryLongQueue{
		lgr: lgr, dir: dir, segSize: segSize,
		prodSubs: []chan bool{},
		pEntries: &pEntries, cEntries: &cEntries,
	}, nil
}
