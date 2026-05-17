package s3meta

import (
	"fmt"
	"os"
	"time"

	"github.com/t-beigbeder/vdasync/dssagrpc"
)

func (s3m *s3Meta) InitRepo() error {
	if err := s3m.initS3Client(); err != nil {
		return err
	}
	dp := s3m.rootPrefix + "/dirs/"
	for _, dir := range []string{".", ".."} {
		key := dp + dir
		exists, err := s3m.repoClient().ObjectExists(key)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("InitBucket: key %s already exists", key)
		}
	}

	if err := s3m.putProtoMessage(dp+".", &dssagrpc.DataEntries{Entries: []*dssagrpc.DataEntry{}}); err != nil {
		return err
	}
	dotDe := dssagrpc.DataEntry{
		IsDir:      true,
		Path:       "/",
		Mtime:      time.Now().Unix(),
		User:       int32(os.Getuid()),
		UserRights: &dssagrpc.Rights{Read: true, Write: true, Execute: true},
		Group:      int32(os.Getgid()),
	}
	if err := s3m.putProtoMessage(dp+"..",
		&dssagrpc.DataEntries{Entries: []*dssagrpc.DataEntry{&dotDe}},
	); err != nil {
		return err
	}
	return nil
}

func (s3m *s3Meta) DeleteRepo() error {
	if err := s3m.initS3Client(); err != nil {
		return err
	}
	return s3m.repoClient().DeleteAll(s3m.rootPrefix + "/")
}
