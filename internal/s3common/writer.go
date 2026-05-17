package s3common

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

type s3Reader struct {
	pull   chan []byte
	buffer []byte
}

func (rdr *s3Reader) Read(p []byte) (n int, err error) {
	if rdr.buffer == nil {
		rdr.buffer = <- rdr.pull
	}
	if len(rdr.buffer) > 0 {
		n := copy(p, rdr.buffer)
		if n == len(rdr.buffer) {
			rdr.buffer = nil
		} else {
			rdr.buffer = rdr.buffer[n:]
		}
		return n, nil
	}
	return 0, io.EOF
}

type ApiWriter struct {
	Key    string
	Rc     *S3RepoClient
	rdr    *s3Reader
	wg     sync.WaitGroup
	push   chan []byte
	rdrErr error
	closed bool
}

func (sw *ApiWriter) Write(p []byte) (int, error) {
	if sw.rdr == nil {
		sw.push = make(chan []byte)
		sw.rdr = &s3Reader{pull: sw.push}
		sw.wg.Add(1)
		go func() {
			sw.rdrErr = sw.Rc.UploadObject(sw.Key, sw.rdr)
			sw.wg.Done()
		}()
	}
	sw.push <- p
	return len(p), nil
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
	close(sw.push)
	sw.wg.Wait()
	if sw.rdrErr != nil {
		return sw.rdrErr
	}
	return nil
}
