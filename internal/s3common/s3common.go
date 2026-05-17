package s3common

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3RepoClient struct {
	Client     *s3.Client
	BucketName string
}

func InitS3Client(profileName string) (cfg aws.Config, client *s3.Client, err error) {
	ctx := context.TODO()
	if profileName == "" {
		cfg, err = config.LoadDefaultConfig(ctx)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profileName))
	}
	if err != nil {
		return
	}
	client = s3.NewFromConfig(cfg)
	return
}

func (rc *S3RepoClient) DeleteAll(prefix string) error {
	ctx := context.TODO()
	for {
		lOO, err := rc.Client.ListObjectsV2(ctx,
			&s3.ListObjectsV2Input{Bucket: &rc.BucketName, Prefix: &prefix},
		)
		if err != nil {
			return err
		}
		if len(lOO.Contents) == 0 {
			return nil
		}
		var oids []types.ObjectIdentifier
		for _, ob := range lOO.Contents {
			oids = append(oids, types.ObjectIdentifier{Key: aws.String(*ob.Key)})
		}
		_, err = rc.Client.DeleteObjects(ctx,
			&s3.DeleteObjectsInput{
				Bucket: &rc.BucketName,
				Delete: &types.Delete{Objects: oids, Quiet: aws.Bool(true)},
			})
		if err != nil {
			return err
		}
	}
}

func (rc *S3RepoClient) ObjectExists(key string) (bool, error) {
	_, err := rc.Client.HeadObject(
		context.TODO(),
		&s3.HeadObjectInput{Bucket: &rc.BucketName, Key: &key},
	)
	if err == nil {
		return true, nil
	}
	var ae *types.NotFound
	if !errors.As(err, &ae) {
		return false, err
	}
	return false, nil
}

func (rc *S3RepoClient) PutObject(key string, data []byte) error {
	bdt := bytes.NewBuffer(data)
	_, err := rc.Client.PutObject(
		context.TODO(),
		&s3.PutObjectInput{Bucket: &rc.BucketName, Key: &key, Body: bdt},
	)
	return err
}

func (rc *S3RepoClient) UploadObject(key string, rdr io.Reader) error {
	//	[profile otvl-tests]
	//	request_checksum_calculation = when_required
	// => no ContentLength required
	_, err := rc.Client.PutObject(
		context.TODO(),
		&s3.PutObjectInput{
			Bucket: &rc.BucketName,
			Key:    &key,
			Body:   rdr},
	)
	return err
}

func (rc *S3RepoClient) GetObject(key string) ([]byte, error) {
	gor, err := rc.Client.GetObject(
		context.TODO(),
		&s3.GetObjectInput{Bucket: &rc.BucketName, Key: &key},
	)
	if err != nil {
		return nil, err
	}
	defer gor.Body.Close()
	return io.ReadAll(gor.Body)
}
