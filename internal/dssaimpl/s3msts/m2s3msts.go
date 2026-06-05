package s3msts

import (
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/metasts"
	"github.com/t-beigbeder/vdasync/internal/s3common"
	"log/slog"
	"path"
)

type m2s3svc struct {
	metasts.M2StSvc
	rootPrefix string
	s3repo     *s3common.S3RepoClient
}

type m2s3StSvc struct {
	msvc *m2s3svc
}

func (m *m2s3StSvc) key() string {
	return path.Join(m.msvc.rootPrefix, "/.vdasync/m2s3msts.meta")
}

// Exists implements [metasts.StorageSvc].
func (m *m2s3StSvc) Exists() (bool, error) {
	return m.msvc.s3repo.ObjectExists(m.key())
}

// Get implements [metasts.StorageSvc].
func (m *m2s3StSvc) Get() ([]byte, error) {
	return m.msvc.s3repo.GetObject(m.key())
}

// Put implements [metasts.StorageSvc].
func (m *m2s3StSvc) Put(bs []byte) error {
	return m.msvc.s3repo.PutObject(m.key(), bs)
}

var _ metasts.StorageSvc = &m2s3StSvc{}

func MakeM2S3MetaStorageSvc(lgr *slog.Logger, profileName, bucketName, rootPrefix string) (metasts.MetaStorageSvc, error) {
	s3repo, err := s3common.NewS3RepoClient(lgr, profileName, bucketName)
	if err != nil {
		return nil, err
	}
	stsvc := &m2s3StSvc{}
	this := &m2s3svc{
		M2StSvc:    metasts.M2StSvc{Lgr: lgr, StSvc: stsvc},
		rootPrefix: rootPrefix,
		s3repo:     s3repo,
	}
	stsvc.msvc = this
	return this, nil
}
