package s3msts

import (
	"os"

	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
)

const (
	testProfile = "otvl-tests"
	testBucket  = "otvl-tests"
)

func getS3Env() (pf, bk, rp string) {
	pf = os.Getenv("OTVL_TEST_S3_PF")
	if pf == "" {
		pf = "otvl-tests"
	}
	bk = os.Getenv("OTVL_TEST_S3_BK")
	if bk == "" {
		bk = "otvl-tests"
	}
	rp = os.Getenv("OTVL_TEST_S3_RP")
	if rp == "" {
		rp = "vdasync/tests/default"
	}
	return
}

func getRepo (t *testing.T) dssa.Dssa {
	pf, bk, rp := getS3Env()
	ds, _, err := MakeS3MstsDssa(pf, bk, rp, MSTS_M2S3)
	require.NoError(t, err)
	return ds
}

func TestXxx(t *testing.T) {
	ds := getRepo(t)
	require.NotNil(t, ds)
}