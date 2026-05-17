package s3meta

import (
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dirindex"
	"github.com/t-beigbeder/vdasync/internal/s3common"
)

type s3Meta struct {
	profileName string
	bucketName  string
	rootPrefix  string
	awsCfg      aws.Config
	s3Client    *s3.Client
	di          dirindex.DirIndex
}

// GetReadCloser implements [dssa.Dssa].
func (s3d *s3Meta) GetReadCloser(string) (io.ReadCloser, error) {
	panic("unimplemented")
}

// GetWriteCloser implements [dssa.Dssa].
func (s3d *s3Meta) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	return &s3common.ApiWriter{
		Key: path.Join(s3d.rootPrefix, "files", path_),
		Rc:  s3d.repoClient(),
		CloseCb: func(nWritten int64, err error) {
			if err != nil {
				return
			}
			s3d.di.Put(&dssa.DataEntry{Path: path_, Size: nWritten, Mtime: time.Now().Unix()})
		},
	}, nil
}

func (s3d *s3Meta) doList(path_ string) ([]*dssa.DataEntry, error) {
	if _, err := s3d.di.Get(path_); err == nil {
		return s3d.di.List(path_)
	}
	gdes := dssagrpc.DataEntries{}
	if err := s3d.getProtoMessage(path.Join(s3d.rootPrefix, path_), &gdes); err != nil {
		return nil, err
	}
	if path_ != "/" {
		pp := path.Dir(path_)
		if _, err := s3d.doList(pp); err != nil {
			return nil, err
		}
	}
	for _, gde := range gdes.Entries {
		de := common.GrpcDte2DssDte(gde)
		if err := s3d.di.Put(de); err != nil {
			return nil, err
		}
	}
	return s3d.di.List(path_)
}

// List implements [dssa.Dssa].
func (s3d *s3Meta) List(path_ string) ([]*dssa.DataEntry, error) {
	if err := s3d.initS3Client(); err != nil {
		return nil, err
	}
	return s3d.doList(path_)
}

// Mkdir implements [dssa.Dssa].
func (s3d *s3Meta) Mkdir(*dssa.DataEntry) error {
	panic("unimplemented")
}

// Rm implements [dssa.Dssa].
func (s3d *s3Meta) Rm(string) error {
	panic("unimplemented")
}

// SetStat implements [dssa.Dssa].
func (s3d *s3Meta) SetStat(_ *dssa.DataEntry, noPerm bool, noMtime bool) error {
	panic("unimplemented")
}

// Stat implements [dssa.Dssa].
func (s3d *s3Meta) Stat(string) (*dssa.DataEntry, error) {
	panic("unimplemented")
}

// Symlink implements [dssa.Dssa].
func (s3d *s3Meta) Symlink(old string, new_ string) error {
	panic("unimplemented")
}

func MakeS3MetaDssa(profileName, bucketName, rootPrefix string) dssa.Dssa {
	return &s3Meta{
		profileName: profileName,
		bucketName:  bucketName,
		rootPrefix:  rootPrefix,
		di:          dirindex.MakeMemIndexDssa(),
	}
}
