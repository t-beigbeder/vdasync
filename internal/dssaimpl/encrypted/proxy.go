package encrypted

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

type SessionMonitor interface {
	WritingServed()
	SomethingServed()
}

type ProxyDssa interface {
	dssa.Dssa
	GetValueSetCb() func(string, []byte) error
	GetEncryptedDssa() EncryptedDssa
	StopSessionMonitor()
}

type proxyDss struct {
	lgr                 *slog.Logger
	rootPath            string
	ageIdentitiesGetter func() []string
	dss                 EncryptedDssa
	mx                  sync.Mutex
	sidGetter           func() string
	recs                []string
	sm                  *sessionMon
}

// StopSessionMonitor implements [ProxyDssa].
func (p *proxyDss) StopSessionMonitor() {
	p.mx.Lock()
	defer p.mx.Unlock()
	if p.dss == nil {
		return
	}
	p.sm.stop()
	p.sm = nil
	p.dss = nil
}

// GetEncryptedDssa implements [ProxyDssa].
func (p *proxyDss) GetEncryptedDssa() EncryptedDssa {
	p.mx.Lock()
	defer p.mx.Unlock()
	return p.dss
}

// GetValueSetCb implements [ProxyDssa].
func (p *proxyDss) GetValueSetCb() func(string, []byte) error {
	return p.setValue
}

const (
	KeyIds    = "identities"
	KeyRecs   = "recipients"
	KeyOpen   = "open"
	KeyClose  = "close"
	KeyRepair = "repair"
)

func (p *proxyDss) openOrRepair(repair bool) error {
	if p.dss != nil {
		return errors.New("encrypted.proxyDss.setValue: already opened")
	}
	dss, err := MakeEncryptedDssa(p.lgr, localfiles.MakeLocalFilesDssa(), p.rootPath,
		strings.Split(p.sidGetter(), ","), p.recs, repair,
	)
	if err != nil {
		return err
	}
	if repair {
		if err = dss.NewSession(); err != nil {
			return err
		}
		err = CheckIndex(p.lgr, localfiles.MakeLocalFilesDssa(), p.rootPath,
			strings.Split(p.sidGetter(), ","), p.recs, true)
		return err
	}
	if err = dss.NewSession(); err != nil {
		return err
	}
	p.dss = dss
	p.sm = newSessionMon(p.lgr, dss)
	return nil
}

func (p *proxyDss) setValue(key string, val []byte) error {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.lgr.Debug("setValue", "key", key)
	switch key {
	case KeyIds:
		dVal, err := common.AgeDecryptMsg(val, p.ageIdentitiesGetter()...)
		if err != nil {
			return err
		}
		p.sidGetter = func() string { return string(dVal) }
	case KeyRecs:
		p.recs = strings.Split(string(val), ",")
	case KeyOpen:
		return p.openOrRepair(false)
	case KeyRepair:
		return p.openOrRepair(true)
	case KeyClose:
		if p.dss == nil {
			return errors.New("encrypted.proxyDss.setValue: already closed")
		}
		p.sm.stop()
		p.sm = nil
		p.dss = nil
	default:
		return fmt.Errorf("encrypted.proxyDss.setValue: unknown key %s", key)
	}
	return nil
}

func (p *proxyDss) checkProxied() (dssa.Dssa, error) {
	p.mx.Lock()
	defer p.mx.Unlock()
	if p.dss == nil {
		return nil, errors.New("encrypted.proxyDss: not configured yet")
	}
	return p.dss, nil
}

// Checksum implements [dssa.Dssa].
func (p *proxyDss) Checksum(algos string, path_ string) (string, error) {
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return "", err
	}
	return dss.Checksum(algos, path_)
}

// EndSession implements [dssa.Dssa].
func (p *proxyDss) EndSession() error {
	p.mx.Lock()
	defer p.mx.Unlock()
	if p.dss == nil {
		return errors.New("encrypted.proxyDss: not configured yet")
	}
	p.sm.endSession()
	return nil
}

// GetReadCloser implements [dssa.Dssa].
func (p *proxyDss) GetReadCloser(path_ string) (io.ReadCloser, error) {
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return nil, err
	}
	return dss.GetReadCloser(path_)
}

// GetWriteCloser implements [dssa.Dssa].
func (p *proxyDss) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return nil, err
	}
	return dss.GetWriteCloser(path_)
}

// List implements [dssa.Dssa].
func (p *proxyDss) List(path_ string) ([]*dssa.DataEntry, error) {
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return nil, err
	}
	return dss.List(path_)
}

// Mkdir implements [dssa.Dssa].
func (p *proxyDss) Mkdir(de *dssa.DataEntry) error {
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return err
	}
	return dss.Mkdir(de)
}

// NewSession implements [dssa.Dssa].
func (p *proxyDss) NewSession() error {
	if _, err := p.checkProxied(); err != nil {
		return err
	}
	return nil
}

// Rm implements [dssa.Dssa].
func (p *proxyDss) Rm(path_ string) error {
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return err
	}
	return dss.Rm(path_)
}

// SetStat implements [dssa.Dssa].
func (p *proxyDss) SetStat(de *dssa.DataEntry, noPerm bool, noMtime bool) error {
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return err
	}
	return dss.SetStat(de, noPerm, noMtime)
}

// Stat implements [dssa.Dssa].
func (p *proxyDss) Stat(path_ string) (*dssa.DataEntry, error) {
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return nil, err
	}
	return dss.Stat(path_)
}

// Symlink implements [dssa.Dssa].
func (p *proxyDss) Symlink(old string, new_ string) error {
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return err
	}
	return dss.Symlink(old, new_)
}

var _ ProxyDssa = &proxyDss{}

func MakeProxyDssa(
	lgr *slog.Logger,
	rootPath string,
	ageIdentities []string,
) (ProxyDssa, error) {
	pxd := &proxyDss{
		lgr:                 lgr,
		rootPath:            rootPath,
		ageIdentitiesGetter: func() []string { return ageIdentities },
	}
	return pxd, nil
}

type sessionMon struct {
	lgr          *slog.Logger
	dss          EncryptedDssa
	served       chan bool
	writing      chan bool
	endsession   chan bool
	done         chan bool
	synchronized chan bool
}

func newSessionMon(lgr *slog.Logger, dss EncryptedDssa) *sessionMon {
	sm := &sessionMon{lgr: lgr, dss: dss}
	sm.served = make(chan bool, 1024)
	sm.writing = make(chan bool, 1024)
	sm.endsession = make(chan bool, 8)
	sm.done = make(chan bool)
	sm.synchronized = make(chan bool)
	dss.SetSessionMonitor(sm)
	go sm.monitors()
	return sm
}

func (sm *sessionMon) runEndSession(final bool) {
	sm.lgr.Debug("sessionMon.runEndSession", "final", final)
	if final {
		if err := sm.dss.EndSession(); err != nil {
			sm.lgr.Error("sessionMon.runEndSession: EndSession", "err", err)
		}
		return
	}
	if err := sm.dss.Msts().SaveSession(); err != nil {
		sm.lgr.Error("sessionMon.runEndSession: SaveSession", "err", err)
	}
}

func (sm *sessionMon) monitors() {
	timer := time.NewTimer(10 * time.Second)
	writings := 0
	for {
		select {
		case <-sm.served:
			timer.Reset(10 * time.Second)
		case <-timer.C:
		case <-sm.endsession:
			timer.Reset(10 * time.Second)
			sm.runEndSession(false)
		case <-sm.writing:
			writings++
			if writings >= 1024 {
				writings = 0
				timer.Reset(10 * time.Second)
				sm.runEndSession(false)
			}
		case <-sm.done:
			sm.lgr.Debug("sessionMon.monitors: done")
			timer.Stop()
			sm.runEndSession(true)
			close(sm.synchronized)
			return
		}
	}

}

func (sm *sessionMon) endSession() {
	sm.lgr.Debug("sessionMon.endSession")
	sm.endsession <- true
}

func (sm *sessionMon) stop() {
	sm.lgr.Debug("sessionMon.stop")
	close(sm.done)
	<-sm.synchronized
}

// SomethingServed implements [SessionMonitor].
func (sm *sessionMon) SomethingServed() {
	sm.served <- true
}

// WritingServed implements [SessionMonitor].
func (sm *sessionMon) WritingServed() {
	sm.writing <- true
}
