package common

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgeEncDec(t *testing.T) {
	pub, pri, err := AgeNewKeyPair()
	require.NoError(t, err)
	ebs, err := AgeEncryptMsg([]byte("TestAgeEncDec"), pub)
	require.NoError(t, err)
	dbs, err := AgeDecryptMsg(ebs, pri)
	require.NoError(t, err)
	require.Equal(t, "TestAgeEncDec", string(dbs))
}

func TestAgeEncDecBinaryStream(t *testing.T) {
	pub, pri, err := AgeNewKeyPair()
	require.NoError(t, err)
	td := t.TempDir()
	txtf := path.Join(td, "test.txt")
	MakeTextTestFile(txtf, 1024*1024)
	ef := path.Join(td, "test.txt.enc")
	wr, err := os.Create(ef)
	require.NoError(t, err)
	ewr, err := AgeEncrypt(wr, pub)
	require.NoError(t, err)
	in, err := os.Open(txtf)
	require.NoError(t, err)
	_, err = io.Copy(ewr, in)
	require.NoError(t, err)
	require.NoError(t, ewr.Close())

	in, err = os.Open(ef)
	require.NoError(t, err)
	rr, err := AgeDecrypt(in, pri)
	require.NoError(t, err)
	txtf2 := path.Join(td, "test-decoded.txt")
	wr2, err := os.Create(txtf2)
	require.NoError(t, err)
	_, err = io.Copy(wr2, rr)
	require.NoError(t, err)
	sh1, err := FileSha256(txtf)
	require.NoError(t, err)
	sh2, err := FileSha256(txtf2)
	require.Equal(t, sh1, sh2)
}
