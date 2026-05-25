package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/s3msts"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"google.golang.org/grpc"
)

func RunS3Plugin() {
	var (
		hostFlag      = flag.String("host", "localhost", "host/address to listen, defaults to localhost")
		portFlag      = flag.Int("port", 0, "port to listen")
		nameFlag      = flag.String("name", "", "plugin name")
		typeFlag      = flag.String("type", "", "plugin type")
		s3ProfileFlag = flag.String("s3profile", "", "aws config profile name, default if not specified")
		s3BucketFlag  = flag.String("s3bucket", "", "aws s3 bucket name")
		s3PrefixFlag  = flag.String("s3prefix", "", "aws s3 prefix in the bucket")
		s3PurgeFlag   = flag.Bool("s3purge", false, "don't run the plugin, clean up all s3 objects under the given prefix")
	)
	cf := cli.CommonFlags()
	flag.Parse()
	exe, err := os.Executable()
	if err != nil {
		common.Fatal(nil, fmt.Errorf("os.Executable: %v", err))
	}
	cmd := path.Base(exe)
	lgr, err := common.CliLogger(cmd, *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("path.Base: %s: %v", exe, err))
	}

	if *s3BucketFlag == "" {
		common.Fatal(lgr, errors.New("s3bucket empty"))
	}
	if *s3PrefixFlag == "" {
		common.Fatal(lgr, errors.New("s3prefix empty"))
	}
	dss, err := s3msts.MakeS3MstsDssa(lgr, *s3ProfileFlag, *s3BucketFlag, *s3PrefixFlag, s3msts.MSTS_M2S3)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("s3msts.MakeS3MstsDssa: %s: %v", exe, err))
	}
	lgr.Debug("main", "s3PurgeFlag", *s3PurgeFlag, "S3Repo", dss.S3Repo())
	if *s3PurgeFlag {
		pf := strings.TrimPrefix(*s3PrefixFlag, "/")
		if !strings.HasSuffix(pf, "/") {
			pf += "/"
		}
		if err = dss.S3Repo().DeleteAll(pf); err != nil {
			common.Fatal(lgr, fmt.Errorf("dss.S3Repo().DeleteAll: %v", err))
		}
		os.Exit(0)
	}

	sop, err := cli.GetServerOrPluginTls(cf)
	if err != nil {
		common.Fatal(lgr, err)
	}
	var sops []grpc.ServerOption
	if sop != nil {
		sops = []grpc.ServerOption{sop}
	}

	if err = dss.Msts().NewSession(); err != nil {
		common.Fatal(lgr, err)
	}
	lgr.Info(fmt.Sprintf("%s.main starting", cmd), "name", *nameFlag, "type", *typeFlag, "host", *hostFlag, "port", *portFlag)
	done := make(chan bool)
	cb := func() {
		lgr.Debug("shutdownCb called, closing done")
		close(done)
	}
	_, _, err = remote.RunOpeDssaServer(
		lgr, context.Background(), *hostFlag, *portFlag,
		sops, dss, cb)
	<-done
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("RunOpeDssaServer failed %s", err))
	}
	lgr.Info(fmt.Sprintf("%s.main done", cmd), "host", *hostFlag, "port", *portFlag)
	if err = dss.Msts().EndSession(); err != nil {
		common.Fatal(lgr, err)
	}
}

func main() {
	RunS3Plugin()
}
