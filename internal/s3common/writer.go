package s3common

import (
	"errors"
	"fmt"
)

type s3Reader struct {
	pull   chan []byte
	buffer []byte
}

// Read implements [io.Reader].
func (s *s3Reader) Read(p []byte) (n int, err error) {
	panic("unimplemented")
}

type ApiWriter struct {
	Key    string
	Rc     *S3RepoClient
	rdr    *s3Reader
	push   chan []byte
	closed bool
}

func (sw *ApiWriter) Write(p []byte) (int, error) {
	if sw.rdr == nil {
		sw.push = make(chan []byte)
		sw.rdr = &s3Reader{pull: sw.push}
		sw.Rc.UploadObject(sw.Key, sw.rdr)
	}
	return 0, errors.New("FIXME")
}

func (sw *ApiWriter) Close() error {
	if sw.rdr == nil {
		if _, err := sw.Write(nil); err != nil {
			return fmt.Errorf("ApiWriter.Close: nil write %v", err)
		}
	}
	if sw.closed {
		return errors.New("ApiWriter.Close: already closed")
	}
	sw.closed = true
	return nil
}
