package s3meta
import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)


func (s3m *s3Meta) initS3Client() error {
	var (
		cfg aws.Config
		err error
	)
	if s3m.s3Client == nil {
		if s3m.profileName == "" {
			cfg, err = config.LoadDefaultConfig(context.TODO())
		} else {
			cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(s3m.profileName))
		}
		if err != nil {
			return err
		}
		s3m.awsCfg = cfg
		s3m.s3Client = s3.NewFromConfig(s3m.awsCfg)
	}
	return nil
}
