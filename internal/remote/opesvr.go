package remote

import (
	"context"
	"time"

	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
)

type opeServer struct {
	opegrpc.UnimplementedOpeServer
	grpcServer *grpc.Server
}

func (s *opeServer) Ready(context.Context, *opegrpc.Empty) (*opegrpc.Bool, error) {
	return &opegrpc.Bool{Value: true}, nil
}

func (s *opeServer) Version(context.Context, *opegrpc.Empty) (*opegrpc.Value, error) {
	return &opegrpc.Value{Value: "0.1"}, nil
}

func (s *opeServer) Shutdown(ctx context.Context, v *opegrpc.Value) (*opegrpc.Bool, error) {
	du, err := time.ParseDuration(v.Value)
	if err != nil {
		return nil, err
	}
	go func ()  {
		time.Sleep(du)
		s.grpcServer.Stop()
	}()
	return &opegrpc.Bool{Value: true}, nil
}
