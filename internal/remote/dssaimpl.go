package remote

import (
	"context"
	"errors"
	"strings"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"google.golang.org/grpc"
)

type dssaImpl struct {
	dssagrpc.UnimplementedDataStorageSystemServer
	grpcServer *grpc.Server
	dssa_      dssa.Dssa
}

func osp2dp(path_ string) []string {
	return strings.Split(path_, "/")
}

func dp2gp(dp []string) *dssagrpc.Path {
	return &dssagrpc.Path{Path: dp}
}

func os2gp(path_ string) *dssagrpc.Path {
	return dp2gp(osp2dp(path_))
}

func ddte2gdte(ddte *dssa.DataEntry) *dssagrpc.DataEntry {
	var sErr string
	if ddte.Error != nil {
		sErr = ddte.Error.Error()
	}
	return &dssagrpc.DataEntry{
		IsDir:         ddte.IsDir,
		Path:          &dssagrpc.Path{Path: ddte.Path},
		Size:          ddte.Size,
		Mtime:         ddte.Mtime,
		User:          int32(ddte.User),
		UserRights:    drts2grts(ddte.UserRights),
		Group:         int32(ddte.Group),
		GroupRights:   drts2grts(ddte.GroupRights),
		OtherRights:   drts2grts(ddte.OtherRights),
		IsSymLink:     ddte.IsSymLink,
		SymLinkTarget: ddte.SymLinkTarget,
		Error:         sErr,
	}
}

func drts2grts(drts dssa.Rights) *dssagrpc.Rights {
	return &dssagrpc.Rights{Read: drts.Read, Write: drts.Write, Execute: drts.Execute}
}

func gdte2ddte(gdte *dssagrpc.DataEntry) *dssa.DataEntry {
	var err error
	if gdte.Error != "" {
		err = errors.New(gdte.Error)
	}
	return &dssa.DataEntry{
		IsDir:         gdte.IsDir,
		Path:          gdte.Path.Path,
		Size:          gdte.Size,
		Mtime:         gdte.Mtime,
		User:          int(gdte.User),
		UserRights:    dssa.Rights{Read: gdte.UserRights.Read, Write: gdte.UserRights.Write, Execute: gdte.UserRights.Execute},
		Group:         int(gdte.Group),
		GroupRights:   dssa.Rights{Read: gdte.GroupRights.Read, Write: gdte.GroupRights.Write, Execute: gdte.GroupRights.Execute},
		OtherRights:   dssa.Rights{Read: gdte.OtherRights.Read, Write: gdte.OtherRights.Write, Execute: gdte.OtherRights.Execute},
		IsSymLink:     gdte.IsSymLink,
		SymLinkTarget: gdte.SymLinkTarget,
		Error:         err,
	}
}

// List implements [dssagrpc.DataStorageSystemServer].
func (s *dssaImpl) List(ctx context.Context, gpath *dssagrpc.Path) (*dssagrpc.DataEntries, error) {
	ddtes, err := s.dssa_.List(gpath.Path)
	if err != nil {
		return nil, err
	}
	gdtes := dssagrpc.DataEntries{}
	for _, ddte := range ddtes {
		gdtes.Entries = append(gdtes.Entries, ddte2gdte(ddte))
	}
	return &gdtes, nil
}

func (s *dssaImpl) Stat(ctx context.Context, gpath *dssagrpc.Path) (*dssagrpc.DataEntry, error) {
	ddte, err := s.dssa_.Stat(gpath.Path)
	if err != nil {
		return nil, err
	}
	return ddte2gdte(ddte), nil
}

func (s *dssaImpl) SetStat(ctx context.Context, gdte *dssagrpc.DataEntry) (*dssagrpc.Empty, error) {
	if err := s.dssa_.SetStat(gdte2ddte(gdte)); err != nil {
		return nil, err
	}
	return &dssagrpc.Empty{}, nil
}
