package common

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

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
}

type cssReader struct {
	rdr    io.Reader
	algoss []string
	hs     []hash.Hash
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
