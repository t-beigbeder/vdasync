package opelogimpl

import (
	"fmt"

	"github.com/joncrlsn/dque"
	"github.com/t-beigbeder/vdasync/opelog"
)

type ldq struct {
	dq *dque.DQue
}

type LdqItem struct {
	S string
}

// Close implements [opelog.Queue].
func (lq *ldq) Close() error {
	return lq.dq.Close()
}

// Dequeue implements [opelog.Queue].
func (lq *ldq) Dequeue() (string, error) {
	as, err := lq.dq.DequeueBlock()
	if err != nil {
		return "", err
	}
	is, ok := as.(*LdqItem)
	if !ok {
		return "", fmt.Errorf("Dequeue: incorrect type: %T", as)
	}
	return is.S, nil
}

// Enqueue implements [opelog.Queue].
func (lq *ldq) Enqueue(s string) error {
	return lq.dq.Enqueue(&LdqItem{s})
}

func MakeDQ(path string, segSize int) (opelog.Queue, error) {
	if segSize == 0 {
		segSize = 1000000
	}
	dq, err := dque.New(".lq.dque", path, segSize, func() any { return &LdqItem{} })
	if err != nil {
		return nil, err
	}
	if err := dq.TurboOn(); err != nil {
		return nil, err
	}
	return &ldq{dq: dq}, nil
}
