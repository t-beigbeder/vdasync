package grpcclient

import (
	"errors"
	"fmt"
	"io"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"google.golang.org/grpc"
)

// A writer that implements client streaming
// each writer write() translates to a grpc call to send() after initial put()
type grpcWriter struct {
	gc      *grpcClient
	path_   dssa.Path
	stream  grpc.ClientStreamingClient[dssagrpc.PushedBlock, dssagrpc.Length]
	written int64
	closed  bool
}

var _ io.WriteCloser = &grpcWriter{}

// Write implements [io.WriteCloser].
func (gw *grpcWriter) Write(p []byte) (int, error) {
	var (
		err error
		gp  *dssagrpc.Path
	)
	if gw.stream == nil {
		gw.stream, err = gw.gc.client.Put(gw.gc.ctx)
		if err != nil {
			return 0, err
		}
		gp = common.DssPath2GrpcPath(gw.path_)
	}
	err = gw.stream.Send(&dssagrpc.PushedBlock{Path: gp, Data: p})
	if err != nil {
		return 0, err
	}
	gw.written += int64(len(p))
	return len(p), nil
}

// Close implements [io.WriteCloser].
func (gw *grpcWriter) Close() error {
	if gw.stream == nil {
		return errors.New("grpcWriter.Close: stream was not yet opened")
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
