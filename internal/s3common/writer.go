package s3common

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type s3Reader struct {
	pull   chan []byte
	buffer []byte
	nr     int64
	lgr    *slog.Logger
}

func (rdr *s3Reader) Read(p []byte) (n int, err error) {
	if rdr.buffer == nil {
		rdr.buffer = <-rdr.pull
	}
	if len(rdr.buffer) > 0 {
		n := copy(p, rdr.buffer)
		if n == len(rdr.buffer) {
			rdr.buffer = nil
		} else {
			rdr.buffer = rdr.buffer[n:]
		}
		rdr.nr += int64(n)
		return n, nil
	}
	return 0, io.EOF
}

type CloseCbType func(int64, error)

type ApiWriter struct {
	Key      string
	Rc       *S3RepoClient
	CloseCb  CloseCbType
	Lgr      *slog.Logger
	nWritten int64
	rdr      *s3Reader
	wg       sync.WaitGroup
	push     chan []byte
	rdrErr   error
	closed   bool
}

func (sw *ApiWriter) Write(p []byte) (int, error) {
	if sw.rdr == nil {
		sw.push = make(chan []byte)
		sw.rdr = &s3Reader{pull: sw.push, lgr: sw.Lgr}
		sw.wg.Add(1)
		go func() {
			sw.rdrErr = sw.Rc.UploadObject(sw.Key, sw.rdr)
			sw.wg.Done()
		}()
	}
	buf := make([]byte, len(p))
	copy(buf, p)
	sw.push <- buf
	sw.nWritten += int64(len(p))
	return len(p), nil
}

func (sw *ApiWriter) Close() (rErr error) {
	defer func() {
		if sw.CloseCb != nil {
			sw.CloseCb(sw.nWritten, rErr)
			sw.CloseCb = nil
		}
	}()
	if sw.rdr == nil {
		if _, err := sw.Write(nil); err != nil {
			rErr = fmt.Errorf("ApiWriter.Close: nil write %v", err)
			return
		}
	}
	if sw.closed {
		rErr = errors.New("ApiWriter.Close: already closed")
		return
	}
	sw.closed = true
	close(sw.push)
	sw.wg.Wait()
	if sw.rdrErr != nil {
		rErr = sw.rdrErr
		return
	}
	return
}
