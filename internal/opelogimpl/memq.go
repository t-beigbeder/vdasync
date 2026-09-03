package opelogimpl

import "github.com/t-beigbeder/vdasync/opelog"

type memQ struct {
	cq chan string
}

// Close implements [Queue].
func (mq *memQ) Close() error {
	close(mq.cq)
	return nil
}

// Get implements [Queue].
func (mq *memQ) Get() (string, error) {
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
