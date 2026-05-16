package s3meta

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/t-beigbeder/vdasync/dssa"
	"io"
)

type s3Meta struct {
	profileName string
	bucketName  string
	rootPrefix  string
	awsCfg      aws.Config
	s3Client    *s3.Client
}

// GetReadCloser implements [dssa.Dssa].
func (s *s3Meta) GetReadCloser(string) (io.ReadCloser, error) {
	panic("unimplemented")
}

// GetWriteCloser implements [dssa.Dssa].
func (s *s3Meta) GetWriteCloser(string) (io.WriteCloser, error) {
	panic("unimplemented")
}

// List implements [dssa.Dssa].
func (s *s3Meta) List(string) ([]*dssa.DataEntry, error) {
	panic("unimplemented")
}

// Mkdir implements [dssa.Dssa].
func (s *s3Meta) Mkdir(*dssa.DataEntry) error {
	panic("unimplemented")
}

// Rm implements [dssa.Dssa].
func (s *s3Meta) Rm(string) error {
	panic("unimplemented")
}

// SetStat implements [dssa.Dssa].
func (s *s3Meta) SetStat(_ *dssa.DataEntry, noPerm bool, noMtime bool) error {
	panic("unimplemented")
}

// Stat implements [dssa.Dssa].
func (s *s3Meta) Stat(string) (*dssa.DataEntry, error) {
	panic("unimplemented")
}

// Symlink implements [dssa.Dssa].
func (s *s3Meta) Symlink(old string, new_ string) error {
	panic("unimplemented")
}

func MakeS3MetaDssa(profileName, bucketName, rootPrefix string) dssa.Dssa {
	return &s3Meta{profileName: profileName, bucketName: bucketName, rootPrefix: rootPrefix}
}
