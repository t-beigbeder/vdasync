package remote

import (
	"context"
	"io"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"google.golang.org/grpc"
)

type dssaImpl struct {
	dssagrpc.UnimplementedDataStorageSystemServer
	lgr        *slog.Logger
	grpcServer *grpc.Server
	dssa_      dssa.Dssa
	callStats  chan string
}

func (s *dssaImpl) NewSession(_ context.Context, _ *dssagrpc.Empty) (*dssagrpc.Empty, error) {
	s.lgr.Info("dssaImpl.NewSession")
	if err := s.dssa_.NewSession(); err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}
func (s *dssaImpl) EndSession(_ context.Context, _ *dssagrpc.Empty) (*dssagrpc.Empty, error) {
	s.lgr.Info("dssaImpl.EndSession")
	if err := s.dssa_.EndSession(); err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}

func (s *dssaImpl) List(ctx context.Context, gpath *dssagrpc.Path) (*dssagrpc.DataEntries, error) {
	s.lgr.Debug("dssaImpl.List", "path", gpath.Path)
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
	s.lgr.Debug("dssaImpl.Mkdir", "path", gdte.Path)
	s.callStats <- "Mkdir"
	if err := s.dssa_.Mkdir(common.GrpcDte2DssDte(gdte)); err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}

func (s *dssaImpl) Stat(ctx context.Context, gpath *dssagrpc.Path) (*dssagrpc.DataEntry, error) {
	s.lgr.Debug("dssaImpl.Stat", "path", gpath.Path)
	s.callStats <- "Stat"
	ddte, _ := s.dssa_.Stat(gpath.Path)
	return common.DssDte2GrpcDte(ddte), nil
}

func (s *dssaImpl) SetStat(ctx context.Context, gssde *dssagrpc.SetStatDataEntry) (*dssagrpc.Empty, error) {
	s.lgr.Debug("dssaImpl.SetStat", "path", gssde.DataEntry.Path)
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
		path_   string
		wc      io.WriteCloser
		written int64
		cw      int
		err     error
	)
	s.lgr.Debug("dssaImpl.Put...")
	s.callStats <- "Put"
	for {
		if gpb, err = stream.Recv(); err == io.EOF {
			s.lgr.Debug("dssaImpl.Put: EOF", "path", path_, "wc", wc != nil)
			if wc != nil {
				if err = wc.Close(); err != nil {
					s.lgr.Debug("dssaImpl.Put: wc close err", "path", path_, "err", err)
					return err
				}
			}
			sErr := stream.SendAndClose(&dssagrpc.Length{Length: written})
			s.lgr.Debug("dssaImpl.Put: exiting", "path", path_, "err", sErr)
			wc = nil
			return sErr
		}
		if err != nil {
			return err
		}
		if wc == nil {
			s.lgr.Debug("dssaImpl.Put", "path", gpb.Path)
			path_ = gpb.Path
			if wc, err = s.dssa_.GetWriteCloser(gpb.Path); err != nil {
				return err
			}
			defer func() {
				if wc == nil {
					return
				}
				clErr := wc.Close()
				s.lgr.Debug("dssaImpl.Put: exiting, closing writer", "path", path_, "err", clErr)
			}()
		}
		s.callStats <- "Put.Recv"
		s.lgr.Debug("dssaImpl.Put", "path", path_, "Data", len(gpb.Data))
		if cw, err = wc.Write(gpb.Data); err != nil {
			return err
		}
		written += int64(cw)
	}
}

func (s *dssaImpl) Get(
	gp *dssagrpc.Path, stream grpc.ServerStreamingServer[dssagrpc.PulledBlock]) error {
	s.lgr.Debug("dssaImpl.Get...")
	s.callStats <- "Get"
	rc, err := s.dssa_.GetReadCloser(gp.Path)
	if err != nil {
		return err
	}
	buffer := make([]byte, 32768)
	for {
		n, err := rc.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		s.callStats <- "Get.Send"
		s.lgr.Debug("dssaImpl.Get", "Send", n, "read err", err)
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
	s.lgr.Debug("dssaImpl.Rm", "path", path_)
	s.callStats <- "Rm"
	err := s.dssa_.Rm(path_.Path)
	if err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}

func (s *dssaImpl) Symlink(ctx context.Context, onp *dssagrpc.OldNewPaths) (*dssagrpc.Empty, error) {
	s.lgr.Debug("dssaImpl.Symlink", "path", onp.New_)
	s.callStats <- "Symlink"
	err := s.dssa_.Symlink(onp.Old, onp.New_)
	if err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}
