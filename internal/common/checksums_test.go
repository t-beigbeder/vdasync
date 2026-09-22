package common

import (
	"fmt"
	"io"
	"os"

	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	r1, err := os.Open(ft)
	require.NoError(t, err)
	defer r1.Close()
	cr1, err := NewChecksumsReader(r1, "sha256")
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, cr1)
	require.NoError(t, err)
	require.Equal(t, h1, cr1.Checksums())

	h2, err := FileChecksum(ft, "sha512")
	require.Nil(t, err)
	require.Equal(t, "sha512:109f30d7354f330b30368e861919725ef6affdbafc5854e52ab827ed3469aed866c3365022193477e52d4dabdc957146af22bd2e4f064e656675a659a2e9bb21", h2)

	h3, err := FileChecksum(ft, "sha512,sha256")
	require.Nil(t, err)
	require.Equal(t, h2+","+h1, h3)
	r3, err := os.Open(ft)
	require.NoError(t, err)
	defer r3.Close()
	cr3, err := NewChecksumsReader(r3, "sha512,sha256")
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, cr3)
	require.NoError(t, err)
	require.Equal(t, h3, cr3.Checksums())

	r4, err := os.Open(ft)
	require.NoError(t, err)
	defer r4.Close()
	cr4, err := NewChecksumsReader(r4, "")
	require.NoError(t, err)
	n4, err := io.Copy(io.Discard, cr4)
	require.NoError(t, err)
	require.Equal(t, len(t.Name()), int(n4))
	require.Equal(t, "", cr4.Checksums())

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

func TestAlgosHandling(t *testing.T) {
	require.Equal(t, "a,b,c,d,e", AddAlgos("a,b,c", "d,e"))
	require.Equal(t, "a,b,c", AddAlgos("a,b,c", ""))
	require.Equal(t, "a,b,c", AddAlgos("a,b,c", "c,b"))
	require.Equal(t, "a,b,c", AddAlgos("", "a,b,c"))
	require.Equal(t, 2, len(Css2Map("a:1,b:2")))
	require.Equal(t, "a:1,b:2", FilterCss("a:1,b:2", "a,b"))
	require.Equal(t, "b:2,a:1", FilterCss("a:1,b:2", "b,a"))
	require.Equal(t, "", FilterCss("a:1,b:2", ""))
	require.Equal(t, "", FilterCss("a:1,b:2", "c"))
	require.Equal(t, "a:1,c:3", FilterCss("a:1,b:2,c:3", "a,c"))
	require.Equal(t, "a,c", AlgosFrom("a:1,b2,c:3"))
}

func TestCsString2Bytes(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestCsString2Bytes.dat")
	require.Nil(t, WriteFile(ft, []byte(t.Name())))
	h1, err := FileChecksum(ft, "sha256")
	require.Nil(t, err)
	require.Equal(t, "sha256:0b3b26c3b2e9c20ffa068810ed8badab23d77a423619c698d0ba96787ed83051", h1)
	h2, err := FileChecksum(ft, "md5")
	require.Nil(t, err)
	require.Equal(t, "md5:6ced4125b87379848fd3807d129b3dca", h2)

	bss1, err := Checksums2TypedChecksums(h1)
	require.NoError(t, err)
	h1b, err := TypedChecksums2Checksums(bss1)
	require.NoError(t, err)
	require.Equal(t, h1, h1b)

	bss12, err := Checksums2TypedChecksums(fmt.Sprintf("%s,%s", h1, h2))
	require.NoError(t, err)
	h12b, err := TypedChecksums2Checksums(bss12)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s,%s", h1, h2), h12b)

	bss21, err := Checksums2TypedChecksums(fmt.Sprintf("%s,%s", h2, h1))
	require.NoError(t, err)
	h21b, err := TypedChecksums2Checksums(bss21)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s,%s", h2, h1), h21b)

	rdr, err := os.Open(ft)
	require.NoError(t, err)
	defer rdr.Close()
	cr, err := NewChecksumsReader(rdr, "sha256,md5")
	require.NoError(t, err)
	nr, err := io.Copy(io.Discard, cr)
	require.NoError(t, err)
	require.Equal(t, len(t.Name()), int(nr))
	require.Equal(t, fmt.Sprintf("%s,%s", h1, h2), cr.Checksums())
	require.Equal(t, bss12, cr.TypedChecksums())
}
