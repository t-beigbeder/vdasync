package s3common

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

type s3Reader struct {
	pull   chan []byte
	buffer []byte
	nr int
	lgr *slog.Logger
}

func (rdr *s3Reader) Read(p []byte) (n int, err error) {
	rdr.lgr.Debug("s3Reader.read", "p", len(p), "rdr.buffer", len(rdr.buffer))
	if rdr.buffer == nil {
		rdr.lgr.Debug("s3Reader.read: pull...")
		rdr.buffer = <-rdr.pull
		sb := string(rdr.buffer)
		sbs := strings.Split(sb, "\n")
		_ = sbs
		rdr.lgr.Debug("s3Reader.read", "pulled rdr.buffer", len(rdr.buffer), "sbs[0]", sbs[0], "sbs[-1]", sbs[len(sbs)-1])
	}
	if len(rdr.buffer) > 0 {
		n := copy(p, rdr.buffer)
		if n == len(rdr.buffer) {
			rdr.buffer = nil
		} else {
			rdr.buffer = rdr.buffer[n:]
		}
		rdr.nr += n
		rdr.lgr.Debug("s3Reader.read", "read n=", n)
		return n, nil
	}
	rdr.lgr.Debug("s3Reader.read: EOF.", "read", rdr.nr)
	return 0, io.EOF
}

type CloseCbType func(int64, error)

type ApiWriter struct {
	Key      string
	Rc       *S3RepoClient
	CloseCb  CloseCbType
	Lgr * slog.Logger
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
			sw.Lgr.Debug("ApiWriter.Write: go Upload...")
			sw.rdrErr = sw.Rc.UploadObject(sw.Key, sw.rdr)
			sw.Lgr.Debug("ApiWriter.Write: go wg.Done...")
			sw.wg.Done()
			sw.Lgr.Debug("ApiWriter.Write: go wg.Done return")
		}()
	}
	sw.Lgr.Debug("ApiWriter.Write: push...", "p", len(p))
	sw.push <- p
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
