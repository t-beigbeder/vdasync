package s3msts

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/s3common"
)

func SkipIf(t *testing.T) {
	if s3common.SkipS3() {
		t.Skip("S3 tests are skipped, set OTVL_TEST_S3 and OTVL_TEST_FULL non-empty")
	}
}

func GetS3Env() (pf, bk, rp string) {
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

func GetRepo(t *testing.T) S3DssaWithMsts {
	pf, bk, rp := GetS3Env()
	ds, err := MakeS3MstsDssa(common.GetNullLogger(), pf, bk, rp, MSTS_M2S3)
	require.NoError(t, err)
	return ds
}

func Cleanup(ds S3DssaWithMsts) error {
	return ds.S3Repo().DeleteAll(ds.RootPrefix() + "/")
}
