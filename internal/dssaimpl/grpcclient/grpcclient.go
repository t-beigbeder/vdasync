package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"sync"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/remote"
)

type grpcClient struct {
	mx          sync.Mutex
	trMx        sync.Mutex
	trId        int
	sids        map[int]bool
	lgr         *slog.Logger
	ctx         context.Context
	client      remote.OpeDssaClient
	crossOsSafe bool
}

type GrpcClient interface {
	GetClient() remote.OpeDssaClient
}

var _ GrpcClient = &grpcClient{}

// OpenSession implements [dssa.SessionManager].
func (gc *grpcClient) OpenSession() (int, error) {
	gc.lgr.Debug("grpcClient.NewSession")
	gc.mx.Lock()
	if gc.sids == nil {
		gc.sids = map[int]bool{}
	}
	gc.mx.Unlock()
	sid, err := gc.client.OpenSession(gc.ctx, &dssagrpc.Empty{})
	if err == nil {
		gc.mx.Lock()
		gc.sids[int(sid.Id)] = true
		gc.mx.Unlock()
	}
	return int(sid.Id), err
}

// CloseSession implements [dssa.SessionManager].
func (gc *grpcClient) CloseSession(id int) error {
	gc.lgr.Debug("grpcClient.CloseSession")
	_, err := gc.client.CloseSession(gc.ctx, &dssagrpc.SessionId{Id: int32(id)})
	if err == nil {
		gc.mx.Lock()
		gc.sids[id] = false
		gc.mx.Unlock()
	}
	return err
}

// SessionsNumber implements [dssa.SessionManager].
func (gc *grpcClient) SessionsNumber() (number int) {
	gc.mx.Lock()
	defer gc.mx.Unlock()
	for _, ok := range gc.sids {
		if ok {
			number++
		}
	}
	return
}

// NewSessionTr implements [dssa.SessionManager].
func (gc *grpcClient) NewSessionTr() (err error) {
	gc.trMx.Lock()
	defer gc.trMx.Unlock()
	if gc.trId > 0 {
		return errors.New("grpcClient: transition session already created")
	}
	gc.trId, err = gc.OpenSession()
	return
}

// EndSessionTr implements [dssa.SessionManager].
func (gc *grpcClient) EndSessionTr() error {
	gc.trMx.Lock()
	defer gc.trMx.Unlock()
	if gc.trId == 0 {
		return errors.New("grpcClient: transition session already ended")
	}
	if err := gc.CloseSession(gc.trId); err != nil {
		return err
	}
	gc.trId = 0
	return nil
}

// NewSession implements [dssa.Dssa].
func (gc *grpcClient) NewSession() error {
	return gc.NewSessionTr()
}

// EndSession implements [dssa.Dssa].
func (gc *grpcClient) EndSession() error {
	return gc.EndSessionTr()
}

// IsGrpcClient implements [dssa.Dssa].
func (gc *grpcClient) IsGrpcClient() bool { return true }

// DssInfo implements [dssa.Dssa].
func (gc *grpcClient) DssInfo() string {
	gc.lgr.Debug("grpcClient.DssInfo")
	rdi := ""
	di, err := gc.client.DssInfo(gc.ctx, &dssagrpc.Empty{})
	if err == nil {
		rdi = di.DssInfo
	} else {
		rdi = fmt.Sprintf("remote error %s", err.Error())
	}
	return fmt.Sprintf("gRPC client for %s", rdi)
}

// List implements [dssa.Dssa].
func (gc *grpcClient) List(path_ string) ([]*dssa.DataEntry, error) {
	gc.lgr.Debug("grpcClient.List", "path", path_)
	if err := common.FileNameIsSafe(path.Base(path_), gc.crossOsSafe); err != nil {
		return nil, err
	}
	gds, err := gc.client.List(gc.ctx, &dssagrpc.Path{Path: path_})
	if err != nil {
		return nil, err
	}
	dds := []*dssa.DataEntry{}
	for _, gd := range gds.Entries {
		dde := common.GrpcDte2DssDte(gd)
		if err := common.FileNameIsSafe(path.Base(dde.Path), gc.crossOsSafe); err != nil {
			return nil, err
		}
		dds = append(dds, dde)
	}
	return dds, nil
}

// Mkdir implements [dssa.Dssa].
func (gc *grpcClient) Mkdir(de *dssa.DataEntry) error {
	gc.lgr.Debug("grpcClient.Mkdir", "de", de.Path)
	if err := common.FileNameIsSafe(path.Base(de.Path), gc.crossOsSafe); err != nil {
		return err
	}
	_, err := gc.client.Mkdir(gc.ctx, common.DssDte2GrpcDte(de))
	return err
}

// Stat implements [dssa.Dssa].
func (gc *grpcClient) Stat(path_ string) (*dssa.DataEntry, error) {
	gc.lgr.Debug("grpcClient.Stat", "de", path_)
	if err := common.FileNameIsSafe(path.Base(path_), gc.crossOsSafe); err != nil {
		return nil, err
	}
	gd, err := gc.client.Stat(gc.ctx, &dssagrpc.Path{Path: path_})
	if err != nil {
		return nil, err
	}
	de := common.GrpcDte2DssDte(gd)
	return de, de.Error
}

// SetStat implements [dssa.Dssa].
func (gc *grpcClient) SetStat(ssde *dssa.DataEntry, noPerm, noMtime bool) error {
	gc.lgr.Debug("grpcClient.SetStat", "de", ssde.Path)
	if err := common.FileNameIsSafe(path.Base(ssde.Path), gc.crossOsSafe); err != nil {
		return err
	}
	_, err := gc.client.SetStat(gc.ctx,
		&dssagrpc.SetStatDataEntry{
			DataEntry: common.DssDte2GrpcDte(ssde),
			NoPerm:    noPerm,
			NoMtime:   noMtime,
		})
	return err
}

// Checksum implements [dssa.Dssa].
func (gc *grpcClient) Checksum(algos string, path_ string) (string, error) {
	gc.lgr.Debug("grpcClient.Checksum", "algos", algos, "path", path_)
	gcs, err := gc.client.Checksum(gc.ctx, &dssagrpc.AlgosAndPath{Algos: algos, Path: path_})
	if err != nil {
		return "", err
	}
	return gcs.Checksums, nil
}

// GetReader implements [dssa.Dssa].
func (gc *grpcClient) GetReadCloser(path_ string) (io.ReadCloser, error) {
	gc.lgr.Debug("grpcClient.GetReadCloser", "path", path_)
	if err := common.FileNameIsSafe(path.Base(path_), gc.crossOsSafe); err != nil {
		return nil, err
	}
	return &grpcReader{lgr: gc.lgr, gc: gc, path_: path_}, nil
}

// GetWriter implements [dssa.Dssa].
func (gc *grpcClient) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	gc.lgr.Debug("grpcClient.GetWriteCloser", "path", path_)
	if err := common.FileNameIsSafe(path.Base(path_), gc.crossOsSafe); err != nil {
		return nil, err
	}
	return &grpcWriter{lgr: gc.lgr, path_: path_, gc: gc}, nil
}

// Symlink implements [dssa.Dssa].
func (gc *grpcClient) Rm(path_ string) error {
	gc.lgr.Debug("grpcClient.Rm", "path", path_)
	if err := common.FileNameIsSafe(path.Base(path_), gc.crossOsSafe); err != nil {
		return err
	}
	_, err := gc.client.Rm(gc.ctx, &dssagrpc.Path{Path: path_})
	return err
}

// Symlink implements [dssa.Dssa].
func (gc *grpcClient) Symlink(old string, new_ string) error {
	gc.lgr.Debug("grpcClient.Symlink", "path", new_)
	if err := common.FileNameIsSafe(path.Base(new_), gc.crossOsSafe); err != nil {
		return err
	}
	_, err := gc.client.Symlink(gc.ctx, &dssagrpc.OldNewPaths{Old: old, New_: new_})
	return err
}

// GetClient implements [GrpcClient].
func (gc *grpcClient) GetClient() remote.OpeDssaClient {
	return gc.client
}

func MakeGrpcClient(lgr *slog.Logger, ctx context.Context, client remote.OpeDssaClient) dssa.Dssa {
	return &grpcClient{lgr: lgr, ctx: ctx, client: client, crossOsSafe: false}
}

var _ GrpcClient = &grpcClient{}
var _ dssa.SessionManager = &grpcClient{}
