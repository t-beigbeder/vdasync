package grpc

import "context"

const testHost = "localhost"

func RunGrpcTestServer() (string, context.CancelFunc, error) {
	ctx, cCancel := context.WithCancel(context.Background())
	_ = ctx // FIXME
	var (
		err  error
		port string
	)
	defer func() {
		if err != nil {
			cCancel()
		}
	}()
	return port, cCancel, nil
}
