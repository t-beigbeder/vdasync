package s3meta

import (
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/s3common"
	"google.golang.org/protobuf/proto"
)

func (s3m *s3Meta) repoClient() *s3common.S3RepoClient {
	return &s3common.S3RepoClient{Client: s3m.s3Client, BucketName: s3m.bucketName}
}

func (s3m *s3Meta) initS3Client() error {
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

func (s3m *s3Meta) putProtoMessage(key string, m proto.Message) error {
	bs, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	return s3m.repoClient().PutObject(key, bs)
}
