package icapsftp

import (
	"io"

	"github.com/t-beigbeder/vdasync/internal/dssaimpl/proxy"
)

type pWriterImpl struct {
	ph     proxy.ProxyHandler
	uWrr   io.WriteCloser
	closed bool
}

func (p *pWriterImpl) doClose() error {
	return p.ph.Handle("WClose", &proxy.ProxyHdlArgs{WC: p.uWrr}).Err
}

// Close implements [io.WriteCloser].
func (p *pWriterImpl) Close() error {
	if p.closed {
		return p.doClose()
	}
	p.closed = true
	return p.doClose()
}

// Write implements [io.WriteCloser].
func (p *pWriterImpl) Write(buf []byte) (n int, err error) {
	ress := p.ph.Handle("Write", &proxy.ProxyHdlArgs{WC: p.uWrr, Buf: buf})
	return ress.NRdWr, ress.Err
}

func makePWriter(ph proxy.ProxyHandler, uWrr io.WriteCloser) io.WriteCloser {
	return &pWriterImpl{ph: ph, uWrr: uWrr}
}
