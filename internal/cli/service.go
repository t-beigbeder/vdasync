package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/encrypted"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/grpcclient"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/plugin"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"github.com/t-beigbeder/vdasync/internal/walker"
	"github.com/t-beigbeder/vdasync/opegrpc"
	"google.golang.org/grpc"
)

func RunPlugins(lgr *slog.Logger, cliCfg *config.CliConfig, cf *CommonFlagsType) ([]*plugin.RunningPlugin, error) {
	tab := func(cfg *config.PluginsOptionsType) ([]string, grpc.DialOption, error) {
		dop, err := GetClientPluginTls(cf, cfg)
		if err != nil {
			return nil, nil, err
		}
		return GetPluginTlsOpts(cf, cfg), dop, err
	}
	rps, err := plugin.RunCliConfig(lgr, cliCfg, tab)
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

func GetGrpcClient(lgr *slog.Logger, cf *CommonFlagsType, host string, port int) (dssa.Dssa, error) {
	dst := config.RemoteDataStore(&config.CliConfig{}, host, port)
	copt, err := GetClientServerTls(cf, dst)
	if err != nil {
		return nil, err
	}
	address := fmt.Sprintf("%s:%d", host, port)
	cli, err := remote.CheckServerReadiness(address, copt)
	if err != nil {
		return nil, err
	}
	dss := grpcclient.MakeGrpcClient(lgr, context.Background(), cli)
	if err = dss.NewSession(); err != nil {
		return nil, err
	}
	return dss, nil
}

type ServiceCtx struct {
	Cmd         string
	Dss         dssa.Dssa
	Root        string
	IsRecur     bool
	IsCheck     bool
	IsStat      bool
	CsAlgos     string
	IsSorted    bool
	IsTSorted   bool
	IsNoOwn     bool
	EncIds      []string
	EncRecs     []string
	TrustRecs   []string
	Latency     string
	Count       int
	Concurrency int
	Lgr         *slog.Logger
	OutFile     io.Writer
}

func DoService(sc *ServiceCtx) error {
	switch sc.Cmd {
	case "list":
		return doList(sc)
	case "trust":
		return doTrust(sc)
	case "latency":
		return doLatency(sc)
	case "version":
		return doVersion(sc)
	case "shutdown":
		return doShutdown(sc)
	default:
		return fmt.Errorf("unknown command: %s", sc.Cmd)
	}
}

func doList(sc *ServiceCtx) error {
	rs := map[string]*walker.DoerEntryStatus{}
	if !sc.IsRecur {
		des, err := sc.Dss.List(sc.Root)
		if err != nil {
			return err
		}
		for _, de := range des {
			cs := ""
			if sc.IsCheck && !de.IsDir && !de.IsSymLink {
				cs, err = sc.Dss.Checksum(sc.CsAlgos, de.Path)
				if err != nil {
					return err
				}
			}
			rs[de.Path] = &walker.DoerEntryStatus{DataEntry: de, Checksum: cs}
		}
	} else {
		du, err := time.ParseDuration(sc.Latency)
		if err != nil {
			return err
		}
		wk, err := walker.RecListCs(sc.Lgr, sc.Concurrency, sc.Dss, sc.Root, "dss", sc.IsCheck, sc.CsAlgos, sc.IsStat, du)
		if err != nil {
			return err
		}
		rs = walker.DoerResult(wk)
	}
	rss := []*walker.DoerEntryStatus{}
	for _, doerEs := range rs {
		if sc.IsSorted || sc.IsTSorted {
			rss = append(rss, doerEs)
		} else {
			sc.OutFile.Write([]byte(common.DataEntryList(doerEs.DataEntry, sc.IsNoOwn, doerEs.Checksum) + "\n"))
		}
	}
	if sc.IsSorted {
		slices.SortFunc(rss, func(a, b *walker.DoerEntryStatus) int {
			return strings.Compare(a.DataEntry.Path, b.DataEntry.Path)
		})
	} else if sc.IsTSorted {
		slices.SortFunc(rss, func(a, b *walker.DoerEntryStatus) int {
			i := a.DataEntry.Mtime - b.DataEntry.Mtime
			if i != 0 {
				return int(i)
			}
			return strings.Compare(a.DataEntry.Path, b.DataEntry.Path)
		})
	} else {
		return nil
	}
	for _, doerEs := range rss {
		sc.OutFile.Write([]byte(common.DataEntryList(doerEs.DataEntry, sc.IsNoOwn, doerEs.Checksum) + "\n"))
	}
	return nil
}

func doTrust(sc *ServiceCtx) error {
	gc, underIsGc := sc.Dss.(grpcclient.GrpcClient)
	if !underIsGc {
		return fmt.Errorf("dss type %T is not a gRPC client", sc.Dss)
	}
	if sc.EncIds == nil {
		return errors.New("doTrust: no ageeids")
	}
	if sc.EncRecs == nil {
		return errors.New("doTrust: no ageerecs")
	}
	if sc.TrustRecs == nil {
		return errors.New("doTrust: no agetrecs")
	}
	idss, err := common.AgeEncryptMsg([]byte(strings.Join(sc.EncIds, ",")), sc.TrustRecs...)
	if err != nil {
		return fmt.Errorf("doTrust: encrypting ageeids failed: %v", err)
	}
	bgCtx := context.Background()
	_, err = gc.GetClient().SetValue(bgCtx, &opegrpc.KeyVal{Key: encrypted.KeyIds, Val: idss})
	if err != nil {
		return fmt.Errorf("doTrust: SetValue identities failed %v", err)
	}
	_, err = gc.GetClient().SetValue(bgCtx,
		&opegrpc.KeyVal{Key: encrypted.KeyRecs, Val: []byte(strings.Join(sc.EncRecs, ","))})
	if err != nil {
		return fmt.Errorf("doTrust: SetValue recipients failed %v", err)
	}
	_, err = gc.GetClient().SetValue(bgCtx,
		&opegrpc.KeyVal{Key: encrypted.KeyOpen})
	if err != nil {
		return fmt.Errorf("doTrust: SetValue open failed %v", err)
	}
	return nil
}

func doLatency(sc *ServiceCtx) error {
	gc, underIsGc := sc.Dss.(grpcclient.GrpcClient)
	if !underIsGc {
		return fmt.Errorf("dss type %T is not a gRPC client", sc.Dss)
	}
	var wg sync.WaitGroup
	var limiter chan bool
	if sc.Concurrency > 0 {
		limiter = make(chan bool, sc.Concurrency)
	}
	for i := 0; i < sc.Count; i++ {
		var err error
		if sc.Concurrency <= 0 {
			_, err = gc.GetClient().Latency(context.Background(), &opegrpc.Value{Value: sc.Latency})
		} else {
			limiter <- true
			wg.Add(1)
			go func() {
				gc.GetClient().Latency(context.Background(), &opegrpc.Value{Value: sc.Latency})
				_ = <-limiter
				wg.Done()
			}()
		}
		if err != nil {
			return fmt.Errorf("doLatency: %v", err)
		}
	}
	wg.Wait()
	return nil
}

func doVersion(sc *ServiceCtx) error {
	gc, underIsGc := sc.Dss.(grpcclient.GrpcClient)
	if !underIsGc {
		return fmt.Errorf("dss type %T is not a gRPC client", sc.Dss)
	}
	v, err := gc.GetClient().Version(context.Background(), &opegrpc.Empty{})
	if err != nil {
		return fmt.Errorf("doVersion: %v", err)
	}
	fmt.Println(v.Value)
	return nil
}

func doShutdown(sc *ServiceCtx) error {
	gc, underIsGc := sc.Dss.(grpcclient.GrpcClient)
	if !underIsGc {
		return fmt.Errorf("dss type %T is not a gRPC client", sc.Dss)
	}
	_, err := gc.GetClient().Shutdown(context.Background(), &opegrpc.Value{Value: "100ms"})
	if err != nil {
		return fmt.Errorf("doShutdown: %v", err)
	}
	return nil
}
