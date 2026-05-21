package s3msts

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/metasts"
	"github.com/t-beigbeder/vdasync/internal/s3common"
)

// s3MetaSts implements dssa.Dssa to store data files as s3 objects
// and delegate meta data storage to a MetaStorageSvc
type s3MetaSts struct {
	profileName string
	bucketName  string
	rootPrefix  string
	awsCfg      aws.Config
	s3Client    *s3.Client
	msts        metasts.MetaStorageSvc
}

func (s3m *s3MetaSts) repoClient() *s3common.S3RepoClient {
	return &s3common.S3RepoClient{Client: s3m.s3Client, BucketName: s3m.bucketName}
}

func (s3m *s3MetaSts) initS3Client() error {
	if s3m.s3Client != nil {
		return nil
	}
	cfg, client, err := s3common.InitS3Client(s3m.profileName)
	if err != nil {
		return err
	}
	s3m.awsCfg = cfg
	s3m.s3Client = client
	return nil
}

func (s3m *s3MetaSts) getDe(path_ string) (*dssa.DataEntry, error) {
	ok, err := s3m.msts.Exists(path_)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return s3m.msts.Get(path_)
}

// GetReadCloser implements [dssa.Dssa].
func (s3m *s3MetaSts) GetReadCloser(path_ string) (io.ReadCloser, error) {
	if err := s3m.initS3Client(); err != nil {
		return nil, err
	}
	return s3m.repoClient().GetReadCloser(path.Join(s3m.rootPrefix, path_))
}

// GetWriteCloser implements [dssa.Dssa].
func (s3m *s3MetaSts) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	de, err := s3m.getDe(path_)
	if err != nil {
		return nil, err
	}
	return &s3common.ApiWriter{
		Key: path.Join(s3m.rootPrefix, path_),
		Rc:  s3m.repoClient(),
		CloseCb: func(nWritten int64, err error) {
			if err != nil {
				return
			}
			if de == nil {
				de = &dssa.DataEntry{
					Path: path_,
					Size: nWritten,
					Mtime: time.Now().Unix(),
					User: os.Getuid(),
					UserRights: dssa.Rights{Read: true, Write: true},
				}
			} else {
				de.Size = nWritten
				de.Mtime = time.Now().Unix()
			}
			s3m.msts.Put(de)
		},
	}, nil
}

// List implements [dssa.Dssa].
func (s3m *s3MetaSts) List(path_ string) ([]*dssa.DataEntry, error) {
	return s3m.msts.List(path_)
}

// Mkdir implements [dssa.Dssa].
func (s3m *s3MetaSts) Mkdir(de *dssa.DataEntry) error {
	return s3m.msts.Put(de)
}

// Rm implements [dssa.Dssa].
func (s3m *s3MetaSts) Rm(path_ string) error {
	ok, err := s3m.msts.Exists(path_)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("s3MetaSts.Rm: %s: no such file or directory")
	}
	de, err := s3m.msts.Get(path_)
	if err != nil {
		return err
	}
	if !de.IsDir || !de.IsSymLink {
		if err = s3m.repoClient().DeleteObject(path.Join(s3m.rootPrefix, path_)); err != nil {
			return err
		}
	}
	return s3m.msts.Del(path_)
}

// SetStat implements [dssa.Dssa].
func (s3m *s3MetaSts) SetStat(de *dssa.DataEntry, noPerm bool, noMtime bool) error {
	ede, err := s3m.getDe(de.Path)
	if err != nil {
		return err
	}
	cde := *de
	if ede != nil {
		if noPerm {
			cde.User, cde.UserRights = ede.User, ede.UserRights
			cde.Group, cde.GroupRights = ede.Group, ede.GroupRights
			cde.OtherRights = ede.OtherRights
		}
		if noMtime {
			cde.Mtime = ede.Mtime
		}
	}
	return s3m.msts.Put(&cde)
}

// Stat implements [dssa.Dssa].
func (s3m *s3MetaSts) Stat(path_ string) (*dssa.DataEntry, error) {
	de, err := s3m.getDe(path_)
	if err != nil {
		return nil, err
	}
	if de == nil {
		de = &dssa.DataEntry{Path: path_, Error: fs.ErrNotExist, ErrNotExist: true}
	}
	return de, de.Error
}

// Symlink implements [dssa.Dssa].
func (s3m *s3MetaSts) Symlink(old string, new_ string) error {
	de, err := s3m.getDe(new_)
	if err != nil {
		return err
	}
	if de != nil {
		return fs.ErrExist
	}
	de = &dssa.DataEntry{
		Path: new_,
		IsSymLink: true,
		SymLinkTarget: old,
		Mtime: time.Now().Unix(),
		User: os.Getuid(),
		UserRights: dssa.Rights{Read: true, Write: true, Execute: true},
	}
	return s3m.msts.Put(de)
}

func MakeS3MstsDssa(profileName, bucketName, rootPrefix string) dssa.Dssa {
	return &s3MetaSts{
		profileName: profileName,
		bucketName:  bucketName,
		rootPrefix:  rootPrefix,
	}
}
