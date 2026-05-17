package s3meta

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func getRepo(t *testing.T) dssa.Dssa {
	pf, bk, rp := getS3Env()
	s3m := s3Meta{
		profileName: pf,
		bucketName:  bk,
		rootPrefix:  rp}
	require.NoError(t, s3m.DeleteRepo())
	require.NoError(t, s3m.InitRepo())
	return MakeS3MetaDssa(pf, bk, rp)
}

func TestFileFunctions(t *testing.T) {
	SkipIf(t)
	td1 := t.TempDir()
	ft := path.Join(td1, "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))

	s3d := getRepo(t)
	des, err := s3d.List("/")
	require.NoError(t, err)
	require.Zero(t, len(des))
}
