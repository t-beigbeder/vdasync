package opelogimpl

import (
	"fmt"

	"github.com/joncrlsn/dque"
	"github.com/t-beigbeder/vdasync/opelog"
)

type largeQ struct {
	dq *dque.DQue
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
	s, ok := as.(string)
	if !ok {
		return "", fmt.Errorf("Dequeue: incorrect type: %T", as)
	}
	return s, nil
}

// Enqueue implements [opelog.Queue].
func (lq *largeQ) Enqueue(s string) error {
	return lq.dq.Enqueue(s)
}

func MakeLargeQ(path string) (opelog.Queue, error) {
	dq, err := dque.New(".lq.dque", path, 1000000, func() any { return "" })
	if err != nil {
		return nil, err
	}
	if err := dq.TurboOn(); err != nil {
		return nil, err
	}
	return &largeQ{dq: dq}, nil
}
