package main

import (
	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/sftpc"
)

func main() {
	df := &cli.DssaFactory{}
	df.Register("sftp", &sftpc.DssaMaker{})
	cli.RunSyncCli(df)
}
