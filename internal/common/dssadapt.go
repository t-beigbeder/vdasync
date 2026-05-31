package common

import (
	"errors"
	"strings"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
)

func DssDte2GrpcDte(ddte *dssa.DataEntry) *dssagrpc.DataEntry {
	var sErr string
	if ddte.Error != nil {
		sErr = ddte.Error.Error()
	}
	return &dssagrpc.DataEntry{
		IsDir:         ddte.IsDir,
		Path:          ddte.Path,
		Size:          ddte.Size,
		Mtime:         ddte.Mtime,
		User:          int32(ddte.User),
		UserRights:    DssRights2GrpcRights(ddte.UserRights),
		Group:         int32(ddte.Group),
		GroupRights:   DssRights2GrpcRights(ddte.GroupRights),
		OtherRights:   DssRights2GrpcRights(ddte.OtherRights),
		IsSymLink:     ddte.IsSymLink,
		SymLinkTarget: ddte.SymLinkTarget,
		Error:         sErr,
		ErrNotExist:   ddte.ErrNotExist,
		Id:            ddte.Id,
	}
}

func DssRights2GrpcRights(drts dssa.Rights) *dssagrpc.Rights {
	return &dssagrpc.Rights{Read: drts.Read, Write: drts.Write, Execute: drts.Execute}
}

func GrpcDte2DssDte(gdte *dssagrpc.DataEntry) *dssa.DataEntry {
	var err error
	if gdte.Error != "" {
		err = errors.New(gdte.Error)
	}
	return &dssa.DataEntry{
		IsDir:         gdte.IsDir,
		Path:          gdte.Path,
		Size:          gdte.Size,
		Mtime:         gdte.Mtime,
		User:          int(gdte.User),
		UserRights:    dssa.Rights{Read: gdte.UserRights.Read, Write: gdte.UserRights.Write, Execute: gdte.UserRights.Execute},
		Group:         int(gdte.Group),
		GroupRights:   dssa.Rights{Read: gdte.GroupRights.Read, Write: gdte.GroupRights.Write, Execute: gdte.GroupRights.Execute},
		OtherRights:   dssa.Rights{Read: gdte.OtherRights.Read, Write: gdte.OtherRights.Write, Execute: gdte.OtherRights.Execute},
		NoLStat:       gdte.NoLstat,
		IsSymLink:     gdte.IsSymLink,
		SymLinkTarget: gdte.SymLinkTarget,
		Error:         err,
		ErrNotExist:   gdte.ErrNotExist,
		Id:            gdte.Id,
	}
}

func RelPath(fullPath, rootPath string) string {
	if fullPath == rootPath {
		return ""
	}
	return strings.Replace(fullPath, rootPath+"/", "", 1)
}
