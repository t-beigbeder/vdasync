package encrypted

import (
	"errors"
	"io"
	"log/slog"

	"github.com/t-beigbeder/vdasync/internal/common"
)

type closeCbType func(int64, error)

type eWriterImpl struct {
	lgr *slog.Logger
	tw       io.WriteCloser
	aw       io.WriteCloser
	nWritten int64
	closeCb  closeCbType
	closed   bool
}

// Close implements [io.WriteCloser].
func (e *eWriterImpl) Close() error {
	if e.closed {
		e.lgr.Debug("eWriterImpl.Close: already closed")
		return errors.New("eWriterImpl.Close: already closed")
	}
	e.closed = true
	err := e.aw.Close()
	if e.closeCb != nil {
		e.closeCb(e.nWritten, err)
	}
	e.aw = nil
	tcErr := e.tw.Close()
	if tcErr != nil {
		e.lgr.Debug("eWriterImpl.Close", "err", tcErr)
	}
	e.closeCb = nil
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

func makeEWriter(lgr *slog.Logger, tw io.WriteCloser, cb closeCbType, ageRecipients ...string) (io.WriteCloser, error) {
	aw, err := common.AgeEncrypt(tw, ageRecipients...)
	if err != nil {
		return nil, err
	}
	return &eWriterImpl{lgr: lgr, tw: tw, aw: aw, closeCb: cb}, nil
}
