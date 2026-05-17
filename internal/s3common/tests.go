package s3common

import "os"

func SkipS3() bool {
	if true { // dev mode
		return false
	}
	if os.Getenv("OTVL_TEST_FULL") == "" {
		return true
	}
	if os.Getenv("OTVL_TEST_S3") == "" {
		return true
	}
	return false
}