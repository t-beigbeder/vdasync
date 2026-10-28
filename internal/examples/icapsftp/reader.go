package icapsftp

import (
	"io"

	"github.com/t-beigbeder/vdasync/internal/dssaimpl/proxy"
)

type pReaderImpl struct {
	ph     proxy.ProxyHandler
	uRdr   io.ReadCloser
	closed bool
}

func (p *pReaderImpl) doClose() error {
	return p.ph.Handle("RClose", &proxy.ProxyHdlArgs{RC: p.uRdr}).Err
}

// Close implements [io.ReadCloser].
func (p *pReaderImpl) Close() error {
	if p.closed {
		return p.doClose()
	}
	p.closed = true
	return p.doClose()
}

// Read implements [io.ReadCloser].
func (p *pReaderImpl) Read(buf []byte) (n int, err error) {
	ress := p.ph.Handle("Read", &proxy.ProxyHdlArgs{RC: p.uRdr, Buf: buf})
	return ress.NRdWr, ress.Err
}

func makePReader(ph proxy.ProxyHandler, uRdr io.ReadCloser) io.ReadCloser {
	return &pReaderImpl{ph: ph, uRdr: uRdr}
}
