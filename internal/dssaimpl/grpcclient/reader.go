package grpcclient

import (
	"errors"
	"io"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"google.golang.org/grpc"
)

// a reader that implements client receiving server streaming
// as data received may not fit with the reader's capacity,
type grpcReader struct {
	gc     *grpcClient
	path_  dssa.Path
	stream grpc.ServerStreamingClient[dssagrpc.PulledBlock]
	buffer []byte
	closed bool
}

var _ io.ReadCloser = &grpcReader{}

// Read implements [io.ReadCloser].
func (gr *grpcReader) Read(p []byte) (int, error) {
	var (
		err error
		pb  *dssagrpc.PulledBlock
	)
	if gr.stream == nil {
		gr.stream, err = gr.gc.client.Get(gr.gc.ctx, common.DssPath2GrpcPath(gr.path_))
		if err != nil {
			return 0, err
		}
	}
	for {
		if len(gr.buffer) > 0 {
			n := copy(p, gr.buffer)
			if n == len(gr.buffer) {
				gr.buffer = nil
			} else {
				gr.buffer = gr.buffer[n:]
			}
			return n, nil
		}
		pb, err = gr.stream.Recv()
		if err != nil {
			return 0, err // which can be io.EOF when server sends nil
		}
		gr.buffer = pb.Data
	}
}

// Close implements [io.ReadCloser].
func (gr *grpcReader) Close() error {
	if gr.stream == nil {
		return errors.New("grpcReader.Close: stream was not yet opened")
	}
	if gr.closed {
		return errors.New("grpcReader.Close: already closed")
	}
	gr.closed = true
	_  = gr.stream.CloseSend()
	return nil
}
