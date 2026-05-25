package remote

import (
	"context"
	"io"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"google.golang.org/grpc"
)

type dssaImpl struct {
	dssagrpc.UnimplementedDataStorageSystemServer
	grpcServer *grpc.Server
	dssa_      dssa.Dssa
	callStats  chan string
}

func (s *dssaImpl) NewSession(_ context.Context, _ *dssagrpc.Empty) (*dssagrpc.Empty, error) {
	return &dssagrpc.Empty{}, nil
}
func (s *dssaImpl) EndSession(_ context.Context, _ *dssagrpc.Empty) (*dssagrpc.Empty, error) {
	return &dssagrpc.Empty{}, nil
}

func (s *dssaImpl) List(ctx context.Context, gpath *dssagrpc.Path) (*dssagrpc.DataEntries, error) {
	s.callStats <- "List"
	ddtes, err := s.dssa_.List(gpath.Path)
	if err != nil {
		return nil, err
	}
	gdtes := dssagrpc.DataEntries{}
	for _, ddte := range ddtes {
		gdtes.Entries = append(gdtes.Entries, common.DssDte2GrpcDte(ddte))
	}
	return &gdtes, nil
}

func (s *dssaImpl) Mkdir(ctx context.Context, gdte *dssagrpc.DataEntry) (*dssagrpc.Empty, error) {
	s.callStats <- "Mkdir"
	if err := s.dssa_.Mkdir(common.GrpcDte2DssDte(gdte)); err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}

func (s *dssaImpl) Stat(ctx context.Context, gpath *dssagrpc.Path) (*dssagrpc.DataEntry, error) {
	s.callStats <- "Stat"
	ddte, _ := s.dssa_.Stat(gpath.Path)
	return common.DssDte2GrpcDte(ddte), nil
}

func (s *dssaImpl) SetStat(ctx context.Context, gssde *dssagrpc.SetStatDataEntry) (*dssagrpc.Empty, error) {
	s.callStats <- "SetStat"
	if err := s.dssa_.SetStat(common.GrpcDte2DssDte(gssde.DataEntry), gssde.NoPerm, gssde.NoMtime); err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}

func (s *dssaImpl) Put(stream grpc.ClientStreamingServer[dssagrpc.PushedBlock, dssagrpc.Length]) error {
	// client-side streaming
	// gets a writer to Dssa
	// while recv-ing on the stream, write to Dssa
	var (
		gpb     *dssagrpc.PushedBlock
		wc      io.WriteCloser
		written int64
		cw      int
		err     error
	)
	s.callStats <- "Put"
	for {
		if gpb, err = stream.Recv(); err == io.EOF {
			if wc != nil {
				if err = wc.Close(); err != nil {
					return err
				}
			}
			return stream.SendAndClose(&dssagrpc.Length{Length: written})
		}
		if err != nil {
			return err
		}
		if wc == nil {
			if wc, err = s.dssa_.GetWriteCloser(gpb.Path); err != nil {
				return err
			}
			defer wc.Close()
		}
		s.callStats <- "Put.Recv"
		if cw, err = wc.Write(gpb.Data); err != nil {
			return err
		}
		written += int64(cw)
	}
}

func (s *dssaImpl) Get(
	gp *dssagrpc.Path, stream grpc.ServerStreamingServer[dssagrpc.PulledBlock]) error {
	s.callStats <- "Get"
	rc, err := s.dssa_.GetReadCloser(gp.Path)
	if err != nil {
		return err
	}
	buffer := make([]byte, 8192)
	for {
		n, err := rc.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		s.callStats <- "Get.Send"
		sErr := stream.Send(&dssagrpc.PulledBlock{Data: buffer[0:n]})
		if sErr != nil {
			return sErr
		}
		if err == io.EOF {
			return nil
		}
	}
}

func (s *dssaImpl) Rm(ctx context.Context, path_ *dssagrpc.Path) (*dssagrpc.Empty, error) {
	s.callStats <- "Rm"
	err := s.dssa_.Rm(path_.Path)
	if err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}

func (s *dssaImpl) Symlink(ctx context.Context, onp *dssagrpc.OldNewPaths) (*dssagrpc.Empty, error) {
	s.callStats <- "Symlink"
	err := s.dssa_.Symlink(onp.Old, onp.New_)
	if err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}
