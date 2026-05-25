package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/grpcclient"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/plugin"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"google.golang.org/grpc"
)

func RunPlugins(lgr *slog.Logger, confData string, cf *CommonFlagsType) ([]*plugin.RunningPlugin, error) {
	tab := func(cfg *config.PluginsOptionsType) ([]string, grpc.DialOption, error) {
		dop, err := GetClientPluginTls(cf, cfg)
		if err != nil {
			return nil, nil, err
		}
		return GetPluginTlsOpts(cf, cfg), dop, err
	}
	rps, err := plugin.RunConfData(lgr, confData, tab)
	if err != nil {
		return nil, err
	}
	return rps, nil
}

func SetSignalHandler(lgr *slog.Logger, rps []*plugin.RunningPlugin) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			lgr.Error("main: signal received, preparing to exit", "signal", sig)
			CleanUp(lgr, rps)
			os.Exit(1)
		}
	}()
}

func CleanUp(lgr *slog.Logger, rps []*plugin.RunningPlugin) {
	lgr.Info("CleanUp: plugins Shutdown")
	plugin.Shutdown(rps)
	lgr.Info("CleanUp: plugins WaitFor")
	plugin.WaitFor(rps)
	for _, rp := range rps {
		if rp.Err != nil {
			lgr.Error("main: plugin error", "error", rp.Err)
		}
	}
}

func GetDssAndRootFor(lgr *slog.Logger, cf *CommonFlagsType, cfg *config.CliConfig, isTarget bool, url string, rps []*plugin.RunningPlugin) (dss dssa.Dssa, root string, err error) {
	var (
		pName string
		host  string
		port  int
	)
	sot := "source"
	if isTarget {
		sot = "target"
	}
	pName, host, port, root, err = ParseUrl(url)
	if err != nil {
		return
	}
	if pName == "" && host == "" && port == 0 {
		dss = localfiles.MakeLocalFilesDssa()
		return
	}
	if pName != "" {
		rp := plugin.PluginFor(pName, rps)
		if rp == nil {
			err = fmt.Errorf("%s: url %s: unkown plugin %s", sot, url, pName)
			return
		}
		dss = grpcclient.MakeGrpcClient(lgr, context.Background(), rp.Client)
		if err = dss.NewSession(); err != nil {
			return
		}
		return
	}
	dst := config.RemoteDataStore(cfg, host, port)
	copt, err := GetClientServerTls(cf, dst)
	if err != nil {
		return
	}
	address := fmt.Sprintf("%s:%d", host, port)
	cli, err := remote.CheckServerReadiness(address, copt)
	if err != nil {
		return
	}
	dss = grpcclient.MakeGrpcClient(lgr, context.Background(), cli)
	if err = dss.NewSession(); err != nil {
		return
	}
	return
}
