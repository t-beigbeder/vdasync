package encrypted

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

type ProxyDssa interface {
	dssa.Dssa
	GetValueSetCb() func(string, []byte) error
}

type proxyDss struct {
	lgr                 *slog.Logger
	rootPath            string
	ageIdentitiesGetter func() []string
	dss                 dssa.Dssa
	mx                  sync.Mutex
	sidGetter           func() string
	recs                []string
}

// GetValueSetCb implements [ProxyDssa].
func (p *proxyDss) GetValueSetCb() func(string, []byte) error {
	return p.setValue
}

const (
	KeyIds = "identities"
	KeyRecs = "recipients"
	KeyOpen = "open"
	KeyClose = "close"
)

func (p *proxyDss) setValue(key string, val []byte) error {
	p.mx.Lock()
	defer p.mx.Unlock()
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
		dss, err := MakeEncryptedDssa(p.lgr, localfiles.MakeLocalFilesDssa(), p.rootPath,
			strings.Split(p.sidGetter(), ","), false, p.recs,
		)
		if err != nil {
			return err
		}
		p.dss = dss
	case KeyClose:
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
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return err
	}
	return dss.EndSession()
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
	var (
		dss dssa.Dssa
		err error
	)
	if dss, err = p.checkProxied(); err != nil {
		return err
	}
	return dss.NewSession()
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
