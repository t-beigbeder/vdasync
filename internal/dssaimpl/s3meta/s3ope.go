package s3meta

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/t-beigbeder/vdasync/dssagrpc"
)

func (s3m *s3Meta) InitBucket() error {
	if err := s3m.initS3Client(); err != nil {
		return err
	}
	key := s3m.rootPrefix + "/dirs/."
	_, err := s3m.s3Client.HeadObject(context.TODO(), &s3.HeadObjectInput{Bucket: &s3m.bucketName, Key: &key})
	if err == nil {
		return fmt.Errorf("InitBucket: key %s already exists", key)
	}
	var ae *types.NotFound
	if !errors.As(err, &ae) {
		return err
	}
	el := dssagrpc.DataEntries{Entries: []*dssagrpc.DataEntry{}}
	_ = el
	return nil
}
