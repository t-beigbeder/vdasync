package s3meta

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestFileFunctions(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	td1 := t.TempDir()
	ft := path.Join(td1, "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))
	s3d := MakeS3MetaDssa(testProfile, testBucket, "vdasync/tests/default")
	r777 := dssa.Rights{Read: true, Write: true, Execute: true}
	re := dssa.DataEntry{IsDir: true, Path: "/", UserRights: r777}
	require.NoError(t, s3d.Mkdir(&re))
}
