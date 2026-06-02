package encrypted

import (
	"io"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/metasts"
)

type m2edsvc struct {
	metasts.M2StSvc
}

var _ metasts.MetaStorageSvc = &m2edsvc{}

type m2edsStSvc struct {
	dss           dssa.Dssa
	metaPath      string
	ageIdentities []string
	ageRecipients []string
}

// Exists implements [metasts.StorageSvc].
func (m *m2edsStSvc) Exists() (bool, error) {
	de, err := m.dss.Stat(m.metaPath)
	if de.Error != nil && !de.ErrNotExist {
		return false, err
	}
	return de != nil, nil
}

// Get implements [metasts.StorageSvc].
func (m *m2edsStSvc) Get() ([]byte, error) {
	rr, err := m.dss.GetReadCloser(m.metaPath)
	if err != nil {
		return nil, err
	}
	defer rr.Close()
	ebs, err := io.ReadAll(rr)
	if err != nil {
		return nil, err
	}
	return common.DecryptMsg(ebs, m.ageIdentities...)
}

// Put implements [metasts.StorageSvc].
func (m *m2edsStSvc) Put(bs []byte) error {
	ebs, err := common.EncryptMsg(bs)
	if err != nil {
		return err
	}
	wr, err := m.dss.GetWriteCloser(m.metaPath)
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
