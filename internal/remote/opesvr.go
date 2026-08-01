package remote

import (
	"context"
	"errors"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/opegrpc"
	"google.golang.org/grpc"
)

type opeServer struct {
	opegrpc.UnimplementedOpeServer
	grpcServer *grpc.Server
	shutdownCb func()
	valueSetCb func(key string, value []byte) error
}

func (s *opeServer) Ready(context.Context, *opegrpc.Empty) (*opegrpc.Bool, error) {
	return &opegrpc.Bool{Value: true}, nil
}

func (s *opeServer) Version(context.Context, *opegrpc.Empty) (*opegrpc.Value, error) {
	return &opegrpc.Value{Value: config.GetVersion()}, nil
}

func (s *opeServer) Shutdown(ctx context.Context, v *opegrpc.Value) (*opegrpc.Bool, error) {
	du, err := time.ParseDuration(v.Value)
	if err != nil {
		return nil, err
	}
	go func() {
		time.Sleep(du)
		if s.grpcServer == nil {
			return
		}
		s.grpcServer.Stop()
		if s.shutdownCb != nil {
			s.shutdownCb()
		}
	}()
	return &opegrpc.Bool{Value: true}, nil
}

func (s *opeServer) SetValue(ctx context.Context, kv *opegrpc.KeyVal) (*opegrpc.Empty, error) {
	if s.valueSetCb == nil {
		return nil, errors.New("SetValue on untrusted server")
	}
	if err := s.valueSetCb(kv.Key, kv.Val); err != nil {
		return nil, err
	}
	return &opegrpc.Empty{}, nil
}

func (s *opeServer) Latency(ctx context.Context, v *opegrpc.Value) (*opegrpc.Empty, error) {
	du, err := time.ParseDuration(v.Value)
	if err != nil {
		return nil, err
	}
	time.Sleep(du)
	return &opegrpc.Empty{}, nil
}

var _ opegrpc.OpeServer = &opeServer{}

func NewOpeServer(grpcServer *grpc.Server, shutdownCb func()) opegrpc.OpeServer {
	return &opeServer{grpcServer: grpcServer, shutdownCb: shutdownCb}
}
