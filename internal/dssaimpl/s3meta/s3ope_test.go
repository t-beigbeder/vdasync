package s3meta

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

const (
	testProfile = "otvl-tests"
	testBucket = "otvl-tests"
)

func TestJustToSee(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	dc, err := config.LoadDefaultConfig(context.TODO())
	require.NoError(t, err)
	_ = dc
	tc, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("otvl-tests"))
	require.NoError(t, err)
	client := s3.NewFromConfig(tc)
    output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
        Bucket: aws.String("otvl-tests"),
    })
	require.NoError(t, err)
	lgr := common.GetLogger()
    for _, object := range output.Contents {
		lgr.Info("list", "key", aws.ToString(object.Key), "size", *object.Size)
    }
}

func TestInit(t *testing.T) {
	dss := MakeS3MetaDssa(testProfile, testBucket, "vdasync/tests/default")
	s3m, ok := dss.(*s3Meta)
	require.True(t, ok)
	err := s3m.InitBucket()
	require.NoError(t, err)
}