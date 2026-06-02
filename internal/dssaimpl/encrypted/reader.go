package encrypted

import (
	"io"

	"github.com/t-beigbeder/vdasync/internal/common"
)

type eReaderImpl struct {
	sr     io.ReadCloser
	ar     io.Reader
	closed bool
}

// Close implements [io.ReadCloser].
func (e *eReaderImpl) Close() error {
	if e.closed {
		return e.sr.Close()
	}
	e.closed = true
	return e.sr.Close()
}

// Read implements [io.ReadCloser].
func (e *eReaderImpl) Read(p []byte) (int, error) {
	return e.ar.Read(p)
}

func makeEReader(sr io.ReadCloser, ageIdentities ...string) (io.ReadCloser, error) {
	ar, err := common.Decrypt(sr, ageIdentities...)
	if err != nil {
		return nil, err
	}
	return &eReaderImpl{sr: sr, ar: ar}, nil
}
