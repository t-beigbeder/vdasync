package encrypted

import (
	"fmt"
	"io"
	"log/slog"
	"path"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/metasts"
)

type m2edsvc struct {
	metasts.M2StSvc
}

var _ metasts.MetaStorageSvc = &m2edsvc{}

type m2edsStSvc struct {
	lgr                 *slog.Logger
	dss                 dssa.Dssa
	rootPath            string
	ageIdentitiesGetter func() []string
	ageRecipients       []string
}

func (m *m2edsStSvc) flagPath() string {
	return path.Join(m.rootPath, ".vdasync.flag")
}

// FlagCreate implements [metasts.StorageSvc].
func (m *m2edsStSvc) FlagCreate() error {
	wc, err := m.dss.GetWriteCloser(m.flagPath())
	if err != nil {
		return err
	}
	defer wc.Close()
	if _, err = wc.Write([]byte("")); err != nil {
		return err
	}
	return wc.Close()
}

// FlagExists implements [metasts.StorageSvc].
func (m *m2edsStSvc) FlagExists() (bool, error) {
	de, err := m.dss.Stat(m.flagPath())
	if err == nil {
		return true, nil
	}
	if de != nil && de.ErrNotExist {
		return false, nil
	}
	return false, err
}

// FlagRemove implements [metasts.StorageSvc].
func (m *m2edsStSvc) FlagRemove() error {
	return m.dss.Rm(m.flagPath())
}

func (m *m2edsStSvc) metaPath() string {
	return path.Join(m.rootPath, ".vdasync.meta")
}

// Exists implements [metasts.StorageSvc].
func (m *m2edsStSvc) Exists() (bool, error) {
	de, err := m.dss.Stat(m.metaPath())
	if de.Error != nil && !de.ErrNotExist {
		return false, err
	}
	return !de.ErrNotExist, nil
}

// Get implements [metasts.StorageSvc].
func (m *m2edsStSvc) Get() ([]byte, error) {
	metaPath := m.metaPath()
	m.lgr.Debug("m2edsStSvc.Get", "metaPath", metaPath)
	rr, err := m.dss.GetReadCloser(metaPath)
	if err != nil {
		return nil, err
	}
	defer rr.Close()
	ebs, err := io.ReadAll(rr)
	if err != nil {
		return nil, err
	}
	return common.AgeDecryptMsg(ebs, m.ageIdentitiesGetter()...)
}

// Put implements [metasts.StorageSvc].
func (m *m2edsStSvc) Put(bs []byte) error {
	metaPath := m.metaPath()
	m.lgr.Debug("m2edsStSvc.Put", "metaPath", metaPath)
	de, err := m.dss.Stat(metaPath)
	if de.Error != nil && !de.ErrNotExist {
		return err
	}
	if !de.ErrNotExist {
		sf := time.Unix(de.Mtime, 0).Format(time.RFC3339)
		if err = common.CopyEntry(m.dss, m.metaPath(), fmt.Sprintf("%s.%s", m.metaPath(), sf)); err != nil {
			return err
		}
	}
	ebs, err := common.AgeEncryptMsg(bs, m.ageRecipients...)
	if err != nil {
		return err
	}
	wr, err := m.dss.GetWriteCloser(metaPath)
	if err != nil {
		return err
	}
	defer wr.Close()
	if _, err = wr.Write(ebs); err != nil {
		return err
	}
	return nil
}

var _ metasts.StorageSvc = &m2edsStSvc{}
