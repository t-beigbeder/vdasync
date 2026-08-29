package opelogimpl

import (
	"fmt"

	"github.com/joncrlsn/dque"
	"github.com/t-beigbeder/vdasync/opelog"
)

type largeQ struct {
	dq *dque.DQue
}

type Item struct {
	S string
}

// Close implements [opelog.Queue].
func (lq *largeQ) Close() error {
	return lq.dq.Close()
}

// Dequeue implements [opelog.Queue].
func (lq *largeQ) Dequeue() (string, error) {
	as, err := lq.dq.DequeueBlock()
	if err != nil {
		return "", err
	}
	is, ok := as.(*Item)
	if !ok {
		return "", fmt.Errorf("Dequeue: incorrect type: %T", as)
	}
	return is.S, nil
}

// Enqueue implements [opelog.Queue].
func (lq *largeQ) Enqueue(s string) error {
	return lq.dq.Enqueue(&Item{s})
}

func MakeLargeQ(path string, segSize int) (opelog.Queue, error) {
	if segSize == 0 {
		segSize = 1000000
	}
	dq, err := dque.New(".lq.dque", path, segSize, func() any { return &Item{} })
	if err != nil {
		return nil, err
	}
	if err := dq.TurboOn(); err != nil {
		return nil, err
	}
	return &largeQ{dq: dq}, nil
}
