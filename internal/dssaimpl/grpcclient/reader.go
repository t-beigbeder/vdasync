package grpcclient

import (
	"io"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"google.golang.org/grpc"
)

// a reader that implements client receiving server streaming
// as data received may not fit with the reader's capacity,
// a buffered reader is plugged
type grpcReader struct {
	gc     *grpcClient
	path_  dssa.Path
	size   int
	stream grpc.ServerStreamingClient[dssagrpc.PulledBlock]
}

var _ io.ReadCloser = &grpcReader{}

// Read implements [io.ReadCloser].
func (gr *grpcReader) Read(p []byte) (n int, err error) {
	var (
		err error
		pb *dssagrpc.PulledBlock
	)
	if gr.stream == nil {
		gr.stream, err = gr.gc.client.Get(gr.gc.ctx,
			&dssagrpc.PathAndSize{Path: gr.path_, Size: int64(gr.size)})
		if err != nil {
			return 0, err
		}
	}
	for {
		pb, err = gr.stream.Recv()
		pb.Data
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
}

// Close implements [io.ReadCloser].
func (gr *grpcReader) Close() error {
	panic("unimplemented")
}
