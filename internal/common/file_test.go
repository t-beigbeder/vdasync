package common

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
)

func TestFileFunctions(t *testing.T) {
	var (
		sz  int64
		err error
		bs  []byte
	)
	ft := path.Join(t.TempDir(), "TestFileFunctions.dat")
	require.False(t, FileExists(ft))
	require.Nil(t, WriteFile(ft, []byte(t.Name())))
	sz, err = FileSize(ft)
	require.Nil(t, err)
	require.Equal(t, len(t.Name()), int(sz))
	bs, err = LoadFile(ft)
	require.Nil(t, err)
	require.Equal(t, []byte(t.Name()), bs)
	var bs2 = [MaxLoadFileSize + 1]byte{}
	bs = bs2[:]
	require.Nil(t, WriteFile(ft, bs))
	sz, err = FileSize(ft)
	require.Nil(t, err)
	require.Equal(t, MaxLoadFileSize+1, int(sz))
	bs, err = LoadFile(ft)
	require.NotNil(t, err)
}

func TestAccessRights(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestAccessRights.dat")
	require.Nil(t, WriteFile(ft, []byte(t.Name())))
	fi, ugIds, ugoRights, err := GetFileStat(ft)
	require.Nil(t, err)
	require.Equal(t, len(t.Name()), int(fi.Size()))
	ugoRights[1] = dssa.Rights{Read: true}
	ugoRights[2] = dssa.Rights{}
	err = SetAccessRights(ft, ugIds, ugoRights)
	require.Nil(t, err)
	fi, ugIds, ugoRights, err = GetFileStat(ft)
	require.Nil(t, err)
	require.False(t, ugoRights[1].Write)
	require.True(t, ugoRights[1].Read)
	mode := Rights2Mod(ugoRights)
	require.Equal(t, mode, fs.FileMode(0640))
}

func TestSha256(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestSha256.dat")
	require.Nil(t, WriteFile(ft, []byte(t.Name())))
	h, err := FileSha256(ft)
	require.Nil(t, err)
	require.Equal(t, "f2a2e3a8f52eccf22084cf440466ca4d00b2203df70fd57b11a408567e5a03ff", h)
	h, err = FileChecksum(ft, "sha256")
	require.Nil(t, err)
	require.Equal(t, "sha256:f2a2e3a8f52eccf22084cf440466ca4d00b2203df70fd57b11a408567e5a03ff", h)
}

func TestChecksum(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestChecksum.dat")
	require.Nil(t, WriteFile(ft, []byte(t.Name())))
	h1, err := FileChecksum(ft, "sha256")
	require.Nil(t, err)
	require.Equal(t, "sha256:4b86be7f5fe5776cd535cdf1e81fdd77c204df48c751f61c121b3e72f6767e1e", h1)
	h2, err := FileChecksum(ft, "sha512")
	require.Nil(t, err)
	require.Equal(t, "sha512:109f30d7354f330b30368e861919725ef6affdbafc5854e52ab827ed3469aed866c3365022193477e52d4dabdc957146af22bd2e4f064e656675a659a2e9bb21", h2)
	h3, err := FileChecksum(ft, "sha512,sha256")
	require.Nil(t, err)
	require.Equal(t, h2+","+h1, h3)
	rs := []string{}
	for i := 0; i < 32; i++ {
		require.Nil(t, WriteFile(ft, fmt.Appendf(nil, "TestChecksum-%02d", i)))
		hl1, err := FileChecksum(ft, "sha256")
		require.Nil(t, err)
		rs = append(rs, hl1)
		require.Equal(t, 64+7, len(hl1))
		hl2, err := FileChecksum(ft, "sha512")
		require.Nil(t, err)
		rs = append(rs, hl2)
		require.Equal(t, 128+7, len(hl2))
		hl3, err := FileChecksum(ft, "sha256,sha512")
		require.Nil(t, err)
		rs = append(rs, hl3)
		require.Equal(t, 64+7+1+128+7, len(hl3))
	}
	require.Equal(t, 3*32, len(rs))
}

func TestStdWriter(t *testing.T) {
	ftp := path.Join(t.TempDir(), "TestStdWriter.txt")
	fd0, err := StdWriter(ftp)
	require.NoError(t, err)
	fd0.Write([]byte(ftp + "\n"))
	require.NoError(t, fd0.Close())
	flns, err := FileLines(ftp)
	require.NoError(t, err)
	require.Equal(t, 1, len(flns))
	require.Equal(t, ftp, flns[0])
	fd1, err := StdWriter("stdout")
	require.NoError(t, err)
	require.Equal(t, os.Stdout, fd1)
	fd1.Write([]byte("stdout\n"))
	fd2, err := StdWriter("stderr")
	require.NoError(t, err)
	fd2.Write([]byte("stderr\n"))
	require.Equal(t, os.Stderr, fd2)
}
