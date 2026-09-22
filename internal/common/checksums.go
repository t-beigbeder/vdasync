package common

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

type AlgoCode opeloggrpc.HalgoCode

const (
	HAL_UNSPECIFIED = AlgoCode(opeloggrpc.HalgoCode_HAL_UNSPECIFIED)
	HAL_MD5         = AlgoCode(opeloggrpc.HalgoCode_HAL_MD5)
	HAL_SHA256      = AlgoCode(opeloggrpc.HalgoCode_HAL_SHA256)
	HAL_SHA512      = AlgoCode(opeloggrpc.HalgoCode_HAL_SHA512)
	HAL_SHA3_256    = AlgoCode(opeloggrpc.HalgoCode_HAL_SHA3_256)
	HAL_SHA3_512    = AlgoCode(opeloggrpc.HalgoCode_HAL_SHA3_512)
)

func (ac AlgoCode) String() string {
	switch ac {
	case HAL_MD5:
		return "md5"
	case HAL_SHA256:
		return "sha256"
	case HAL_SHA512:
		return "sha512"
	case HAL_SHA3_256:
		return "sha3_256"
	case HAL_SHA3_512:
		return "sha3_512"
	default:
		return ""
	}
}

func AlgoCodeFor(hName string) AlgoCode {
	switch hName {
	case "md5":
		return HAL_MD5
	case "sha256":
		return HAL_SHA256
	case "sha512":
		return HAL_SHA512
	case "sha3_256":
		return HAL_SHA3_256
	case "sha3_512":
		return HAL_SHA3_512
	default:
		return HAL_UNSPECIFIED
	}
}

func TypedChecksums2Checksums(tcss [][]byte) (string, error) {
	if len(tcss) == 0 {
		return "", nil
	}
	cs := []string{}
	for _, tcs := range tcss {
		if len(tcs) == 0 || AlgoCode(tcs[0]).String() == "" {
			ac := byte(0)
			if len(tcs) > 0 {
				ac = tcs[0]
			}
			return "", fmt.Errorf("bad hash algo code %v", ac)
		}
		fmt_ := fmt.Sprintf("%%0%dx", len(tcs)-1)
		cs = append(cs, fmt.Sprintf("%s:%s", AlgoCode(tcs[0]).String(), fmt.Sprintf(fmt_, tcs[1:])))
	}
	return strings.Join(cs, ","), nil
}

func Checksums2TypedChecksums(hcss string) ([][]byte, error) {
	if hcss == "" {
		return nil, nil
	}
	tcss := [][]byte{}
	for _, hcs := range strings.Split(hcss, ",") {
		hcsSl := strings.Split(hcs, ":")
		if len(hcsSl) != 2 {
			continue
		}
		hName := hcsSl[0]
		scs := hcsSl[1]
		h, err := HashFactory(hName)
		if err != nil {
			return nil, err
		}
		tcs := make([]byte, h.Size()+1)
		tcs[0] = byte(AlgoCodeFor(hName))
		bs, err := hex.DecodeString(scs)
		if err != nil {
			return nil, err
		}
		if len(bs) != h.Size() {
			return nil, fmt.Errorf("checksum size %d for algo %s (%d)", len(bs), hName, h.Size())
		}
		copy(tcs[1:], bs)
		tcss = append(tcss, tcs)
	}
	return tcss, nil
}

func AddAlgos(algos, added string) string {
	if added == "" {
		return algos
	}
	if algos == "" {
		return added
	}
	sAlgos := strings.Split(algos, ",")
	sAdded := strings.Split(added, ",")
	for _, add := range sAdded {
		if slices.Contains(sAlgos, add) {
			continue
		}
		sAlgos = append(sAlgos, add)
	}
	return strings.Join(sAlgos, ",")
}

func Css2Map(css string) map[string]string {
	sCss := strings.Split(css, ",")
	res := make(map[string]string, len(sCss))
	for _, cs := range sCss {
		scs := strings.Split(cs, ":")
		if len(scs) != 2 {
			continue
		}
		res[scs[0]] = scs[1]
	}
	return res
}

func AlgosFrom(css string) string {
	sAlgos := []string{}
	for scs := range strings.SplitSeq(css, ",") {
		ac := strings.Split(scs, ":")
		if len(ac) == 2 {
			sAlgos = append(sAlgos, ac[0])
		}
	}
	return strings.Join(sAlgos, ",")
}

func FilterCss(css, algos string) string {
	if algos == "" {
		return ""
	}
	cm := Css2Map(css)
	sRes := make([]string, 0, len(cm))
	for algo := range strings.SplitSeq(algos, ",") {
		c, ok := cm[algo]
		if ok {
			sRes = append(sRes, fmt.Sprintf("%s:%s", algo, c))
		}
	}
	return strings.Join(sRes, ",")
}

func ReaderSha256(rdr io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, rdr); err != nil {
		return "", err
	}

	return fmt.Sprintf("%064x", h.Sum(nil)), nil
}

func FileSha256(path_ string) (string, error) {
	f, err := os.Open(path_)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return ReaderSha256(f)
}

func HashFactory(hName string) (hash.Hash, error) {
	switch hName {
	case "md5":
		return md5.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	case "sha3_256":
		return sha3.New256(), nil
	case "sha3_512":
		return sha3.New512(), nil
	default:
		return nil, fmt.Errorf("HashFactory: not implemented: %s", hName)
	}
}

func hsFor(algos string) ([]string, []hash.Hash, error) {
	hs := []hash.Hash{}
	if algos == "" {
		algos = "sha256"
	}
	algoss := strings.Split(algos, ",")
	for _, algo := range algoss {
		h, err := HashFactory(algo)
		if err != nil {
			return nil, nil, err
		}
		hs = append(hs, h)
	}
	return algoss, hs, nil
}

func ReaderChecksum(rdr io.Reader, algos string) (string, error) {
	algoss, hs, err := hsFor(algos)
	if err != nil {
		return "", err
	}
	buffer := make([]byte, 32768)
	for {
		n, err := rdr.Read(buffer)
		if err != nil && err != io.EOF {
			return "", err
		}
		for _, h := range hs {
			_, err := h.Write(buffer[0:n])
			if err != nil {
				return "", err
			}
		}
		if err == io.EOF {
			break
		}
	}
	cs := []string{}
	for ix, h := range hs {
		fmt_ := fmt.Sprintf("%%0%dx", h.Size())
		cs = append(cs, fmt.Sprintf("%s:%s", algoss[ix], fmt.Sprintf(fmt_, h.Sum(nil))))
	}
	return strings.Join(cs, ","), nil
}

func FileChecksum(path_ string, algos string) (string, error) {
	f, err := os.Open(path_)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return ReaderChecksum(f, algos)
}

type ChecksumsReader interface {
	io.Reader
	Checksums() string
	// first byte is AlgoCode, checksum comes after
	TypedChecksums() [][]byte
}

type cssReader struct {
	rdr    io.Reader
	algoss []string
	hs     []hash.Hash
}

// TypedChecksums implements [ChecksumsReader].
func (cssr *cssReader) TypedChecksums() [][]byte {
	if len(cssr.algoss) == 0 {
		return nil
	}
	tcss := [][]byte{}
	for ix, h := range cssr.hs {
		tcs := make([]byte, h.Size()+1)
		tcs[0] = byte(AlgoCodeFor(cssr.algoss[ix]))
		copy(tcs[1:], h.Sum(nil))
		tcss = append(tcss, tcs)
	}
	return tcss
}

// Checksums implements [ChecksumsReader].
func (cssr *cssReader) Checksums() string {
	if len(cssr.algoss) == 0 {
		return ""
	}
	cs := []string{}
	for ix, h := range cssr.hs {
		fmt_ := fmt.Sprintf("%%0%dx", h.Size())
		cs = append(cs, fmt.Sprintf("%s:%s", cssr.algoss[ix], fmt.Sprintf(fmt_, h.Sum(nil))))
	}
	return strings.Join(cs, ",")
}

// Read implements [ChecksumsReader].
func (cssr *cssReader) Read(buffer []byte) (n int, err error) {
	n, err = cssr.rdr.Read(buffer)
	if err != nil && err != io.EOF {
		return
	}
	if len(cssr.algoss) == 0 {
		return
	}
	rErr := err
	for _, h := range cssr.hs {
		_, err = h.Write(buffer[0:n])
		if err != nil {
			return
		}
	}
	err = rErr
	return
}

func NewChecksumsReader(rdr io.Reader, algos string) (ChecksumsReader, error) {
	if algos == "" {
		return &cssReader{rdr: rdr}, nil
	}
	algoss, hs, err := hsFor(algos)
	if err != nil {
		return nil, err
	}
	return &cssReader{rdr: rdr, algoss: algoss, hs: hs}, nil
}
