package s3meta

import (
	"testing"

	"github.com/t-beigbeder/vdasync/internal/s3common"
)

func SkipIf(t *testing.T) {
	if s3common.SkipS3() {
		t.Skip("S3 tests are skipped, set OTVL_TEST_S3 and OTVL_TEST_FULL non-empty")
	}
}
