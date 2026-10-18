package proxy

import "io"

type pReaderImpl struct {
	ph     ProxyHandler
	uRdr   io.ReadCloser
	closed bool
}

// Close implements [io.ReadCloser].
func (p *pReaderImpl) Close() error {
	if p.closed {
		return p.uRdr.Close()
	}
	p.closed = true
	return p.uRdr.Close()
}

// Read implements [io.ReadCloser].
func (p *pReaderImpl) Read(buf []byte) (n int, err error) {
	return p.uRdr.Read(buf)
}

func makePReader(ph ProxyHandler, uRdr io.ReadCloser) io.ReadCloser {
	return &pReaderImpl{ph: ph, uRdr: uRdr}
}
