package encrypted

import (
	"errors"
	"io"

	"github.com/t-beigbeder/vdasync/internal/common"
)

type eWriterImpl struct {
	aw     io.WriteCloser
	closed bool
}

// Close implements [io.WriteCloser].
func (e *eWriterImpl) Close() error {
	if e.closed {
		return errors.New("eWriterImpl.Close: already closed")
	}
	e.closed = true
	return e.aw.Close()
}

// Write implements [io.WriteCloser].
func (e *eWriterImpl) Write(p []byte) (n int, err error) {
	return e.aw.Write(p)
}

func makeEWriter(tw io.WriteCloser, ageRecipients ...string) (io.WriteCloser, error) {
	aw, err := common.Encrypt(tw, ageRecipients...)
	if err != nil {
		return nil, err
	}
	return &eWriterImpl{aw: aw}, nil
}
