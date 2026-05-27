package grpcclient

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssagrpc"
	"google.golang.org/grpc"
)

// A writer that implements client streaming
// each writer write() translates to a grpc call to send() after initial put()
type grpcWriter struct {
	lgr     *slog.Logger
	gc      *grpcClient
	path_   string
	stream  grpc.ClientStreamingClient[dssagrpc.PushedBlock, dssagrpc.Length]
	written int64
	closed  bool
}

var _ io.WriteCloser = &grpcWriter{}

// Write implements [io.WriteCloser].
func (gw *grpcWriter) Write(p []byte) (int, error) {
	var (
		pPath string
		err   error
	)
	if gw.stream == nil {
		gw.stream, err = gw.gc.client.Put(gw.gc.ctx)
		if err != nil {
			return 0, err
		}
		pPath = gw.path_
	}
	gw.lgr.Debug("grpcWriter.Write", "path", gw.path_, "p", len(p))
	for start := 0; start <= len(p); {
		if start == len(p) && start > 0 {
			break
		}
		end := len(p)
		if end-start > 8192 {
			end = start + 8192
		}
		gw.lgr.Debug("grpcWriter.Write", "path", gw.path_, "end-start", end-start)
		err = gw.stream.Send(&dssagrpc.PushedBlock{Path: pPath, Data: p[start:end]})
		pPath = ""
		written := end - start
		gw.written += int64(written)
		start += written
		if err != nil {
			return 0, err
		}
		if written == 0 {
			break
		}
	}
	gw.lgr.Debug("grpcWriter.Write: written", "path", gw.path_, "p", len(p), "written", gw.written)
	return len(p), nil
}

// Close implements [io.WriteCloser].
func (gw *grpcWriter) Close() error {
	if gw.stream == nil {
		if _, err := gw.Write(nil); err != nil {
			return fmt.Errorf("grpcWriter.Close: nil write %v", err)
		}
	}
	if gw.closed {
		return errors.New("grpcWriter.Close: already closed")
	}
	gw.lgr.Debug("grpcWriter.Close", "path", gw.path_)
	gw.closed = true
	gl, err := gw.stream.CloseAndRecv()
	if err != nil {
		return err
	}
	if gl.Length != gw.written {
		return fmt.Errorf("grpcWriter.Close: client wrote %d server wrote %d", gl.Length, gw.written)
	}
	return nil
}
