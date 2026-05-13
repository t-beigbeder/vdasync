package grpcclient

import (
	"errors"
	"fmt"
	"io"

	"github.com/t-beigbeder/vdasync/dssagrpc"
	"google.golang.org/grpc"
)

// A writer that implements client streaming
// each writer write() translates to a grpc call to send() after initial put()
type grpcWriter struct {
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
		err error
	)
	if gw.stream == nil {
		gw.stream, err = gw.gc.client.Put(gw.gc.ctx)
		if err != nil {
			return 0, err
		}
		pPath = gw.path_
	}
	for start := 0; start <= len(p); {
		end := len(p)
		if end - start > 8192 {
			end = start + 8192
		}
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
