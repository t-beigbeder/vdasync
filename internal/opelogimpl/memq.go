package opelogimpl

import "github.com/t-beigbeder/vdasync/opelog"

type memQ struct {
	cq chan []byte
}

// Close implements [Queue].
func (mq *memQ) Close() error {
	close(mq.cq)
	return nil
}

// Get implements [Queue].
func (mq *memQ) Get() ([]byte, error) {
	return <-mq.cq, nil
}

// Put implements [Queue].
func (mq *memQ) Put(bs []byte) error {
	mq.cq <- bs
	return nil
}

func NewMemQueue(conc int) opelog.Queue {
	mq := &memQ{
		cq: make(chan []byte, conc+1),
	}
	return mq
}