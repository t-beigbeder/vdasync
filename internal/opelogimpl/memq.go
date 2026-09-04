package opelogimpl

import (
	"errors"
	"sync"

	"github.com/t-beigbeder/vdasync/opelog"
)

type memQ struct {
	cq     chan string
	mx     sync.Mutex
	closed bool
}

// Close implements [Queue].
func (mq *memQ) Close() error {
	mq.mx.Lock()
	defer mq.mx.Unlock()
	if mq.closed {
		return errors.New("memq.Close: already closed")
	}
	mq.closed = true
	return nil
}

// Get implements [Queue].
func (mq *memQ) Get() (string, error) {
	mq.mx.Lock()
	defer mq.mx.Unlock()
	if mq.closed {
		return "", errors.New("memq.Get: all is read on closed queue")
	}
	return <-mq.cq, nil
}

// Put implements [Queue].
func (mq *memQ) Put(s string) error {
	mq.cq <- s
	return nil
}

func NewMemQueue(conc int) opelog.Queue {
	mq := &memQ{
		cq: make(chan string, conc+1),
	}
	return mq
}
