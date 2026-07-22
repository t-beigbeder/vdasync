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

func (p *proxyDss) setValue(key string, eVal []byte) error {
	dVal, err := common.AgeDecryptMsg(eVal, p.ageIdentitiesGetter()...)
	if err != nil {
		return err
	}
	p.mx.Lock()
	defer p.mx.Unlock()
	switch key {
	case "identities":
		p.sidGetter = func() string { return string(dVal) }
	case "recipients":
		p.recs = strings.Split(string(dVal), ",")
	case "open":
		dss, err := MakeEncryptedDssa(p.lgr, localfiles.MakeLocalFilesDssa(), p.rootPath,
			strings.Split(p.sidGetter(), ","), false, p.recs,
		)
		if err != nil {
			return err
		}
		p.dss = dss
	default:
		return fmt.Errorf("encrypted.proxyDss.SetValue: unknown key %s", key)
	}
	return nil
}

func (p *proxyDss) checkProxied() error {
	if p.dss == nil {
		return errors.New("encrypted.proxyDss: not configured yet")
	}
	return nil
}

// Checksum implements [dssa.Dssa].
func (p *proxyDss) Checksum(algos string, path_ string) (string, error) {
	if err := p.checkProxied(); err != nil {
		return "", err
	}
	return p.dss.Checksum(algos, path_)
}

// EndSession implements [dssa.Dssa].
func (p *proxyDss) EndSession() error {
	if err := p.checkProxied(); err != nil {
		return err
	}
	return p.dss.EndSession()
}

// GetReadCloser implements [dssa.Dssa].
func (p *proxyDss) GetReadCloser(path_ string) (io.ReadCloser, error) {
	if err := p.checkProxied(); err != nil {
		return nil, err
	}
	return p.dss.GetReadCloser(path_)
}

// GetWriteCloser implements [dssa.Dssa].
func (p *proxyDss) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	if err := p.checkProxied(); err != nil {
		return nil, err
	}
	return p.dss.GetWriteCloser(path_)
}

// List implements [dssa.Dssa].
func (p *proxyDss) List(path_ string) ([]*dssa.DataEntry, error) {
	if err := p.checkProxied(); err != nil {
		return nil, err
	}
	return p.dss.List(path_)
}

// Mkdir implements [dssa.Dssa].
func (p *proxyDss) Mkdir(de *dssa.DataEntry) error {
	if err := p.checkProxied(); err != nil {
		return err
	}
	return p.dss.Mkdir(de)
}

// NewSession implements [dssa.Dssa].
func (p *proxyDss) NewSession() error {
	if err := p.checkProxied(); err != nil {
		return err
	}
	return p.dss.NewSession()
}

// Rm implements [dssa.Dssa].
func (p *proxyDss) Rm(path_ string) error {
	if err := p.checkProxied(); err != nil {
		return err
	}
	return p.dss.Rm(path_)
}

// SetStat implements [dssa.Dssa].
func (p *proxyDss) SetStat(de *dssa.DataEntry, noPerm bool, noMtime bool) error {
	if err := p.checkProxied(); err != nil {
		return err
	}
	return p.dss.SetStat(de, noPerm, noMtime)
}

// Stat implements [dssa.Dssa].
func (p *proxyDss) Stat(path_ string) (*dssa.DataEntry, error) {
	if err := p.checkProxied(); err != nil {
		return nil, err
	}
	return p.dss.Stat(path_)
}

// Symlink implements [dssa.Dssa].
func (p *proxyDss) Symlink(old string, new_ string) error {
	if err := p.checkProxied(); err != nil {
		return err
	}
	return p.dss.Symlink(old, new_)
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
