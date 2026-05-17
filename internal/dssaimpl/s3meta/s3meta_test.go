package s3meta

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
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
	fn := "/TestFileFunctions.dat"
	s3d := getRepo(t)

	des, err := s3d.List("/")
	require.NoError(t, err)
	require.Zero(t, len(des))

	wc, err := s3d.GetWriteCloser(fn)
	require.NoError(t, err)
	wc.Write([]byte(fn))
	require.NoError(t, wc.Close())
	defer wc.Close()
	des, err = s3d.List("/")
	require.NoError(t, err)
	require.Equal(t, 1, len(des))

	// TODO: flush dirs index to s3
}
