package encrypted

import (
	"errors"
	"io"

	"github.com/t-beigbeder/vdasync/internal/common"
)

type closeCbType func(int64, error)

type eWriterImpl struct {
	aw       io.WriteCloser
	nWritten int64
	closeCb  closeCbType
	closed   bool
}

// Close implements [io.WriteCloser].
func (e *eWriterImpl) Close() error {
	if e.closed {
		return errors.New("eWriterImpl.Close: already closed")
	}
	e.closed = true
	err := e.aw.Close()
	if e.closeCb != nil {
		e.closeCb(e.nWritten, err)
	}
	return err
}

// Write implements [io.WriteCloser].
func (e *eWriterImpl) Write(p []byte) (int, error) {
	n, err := e.aw.Write(p)
	if err != nil {
		return n, err
	}
	e.nWritten += int64(n)
	return n, nil
}

func makeEWriter(tw io.WriteCloser, cb closeCbType, ageRecipients ...string) (io.WriteCloser, error) {
	aw, err := common.AgeEncrypt(tw, ageRecipients...)
	if err != nil {
		return nil, err
	}
	return &eWriterImpl{aw: aw, closeCb: cb}, nil
}
