package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"github.com/t-beigbeder/vdasync/opegrpc"
	"google.golang.org/grpc"
)

type RunningPlugin struct {
	config           *config.CliConfig
	Plugin           *config.PluginType
	port             int
	cmd              *exec.Cmd
	Client           remote.OpeDssaClient
	TlsClientOptions grpc.DialOption
	Err              error
}

type TlsArgsBuilder func(*config.PluginsOptionsType) ([]string, grpc.DialOption, error)

func checkReadiness(rp *RunningPlugin) {
	if rp.config.PluginReadyRetries == 0 || rp.config.PluginReadyTimeout == "" {
		return
	}
	retryDelay, err := time.ParseDuration(rp.config.PluginReadyTimeout)
	if err != nil {
		rp.Err = fmt.Errorf("ParseDuration failed with error %s", err)
		return
	}
	opts := rp.TlsClientOptions
	for count := 0; count < rp.config.PluginReadyRetries; count++ {
		cli, err := remote.CheckServerReadiness(fmt.Sprintf("%s:%d", rp.config.PluginAddress, rp.port), opts)
		rp.Err = err
		if err == nil {
			rp.Client = cli
			break
		}
		time.Sleep(time.Duration(retryDelay))
		retryDelay *= 2
	}
}

func shutdown(rp *RunningPlugin) {
	_, rp.Err = rp.Client.Shutdown(context.Background(), &opegrpc.Value{Value: "100ms"})
}

func waitFor(rp *RunningPlugin) {
	err := rp.cmd.Wait()
	if err != nil {
		rp.Err = fmt.Errorf("child %s wait error %s", rp.cmd, err)
	}
}

func applyIfOK(rps []*RunningPlugin, run func(*RunningPlugin)) {
	for _, rp := range rps {
		if rp.cmd == nil || rp.Err != nil {
			continue
		}
		run(rp)
	}
}

func getExecutablePath(plugin *config.PluginType) (string, error) {
	if plugin.ExecutablePath != "" {
		return plugin.ExecutablePath, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return path.Join(path.Dir(exe), fmt.Sprintf("%s%s", plugin.Type, path.Ext(exe))), nil
}

func RunConfData(lgr *slog.Logger, yamlConf string, tab TlsArgsBuilder) ([]*RunningPlugin, error) {
	lgr.Info("RunConfData: starting")
	config, err := config.Load(yamlConf)
	if err != nil {
		return nil, err
	}
	rps := []*RunningPlugin{}
	for _, plugin := range config.Plugins {
		lgr.Info("RunConfData: starting plugin", "plugin", plugin)
		crp := RunningPlugin{config: config, Plugin: plugin, port: plugin.Port}
		if crp.port == 0 {
			port, err := common.GetFreePort()
			if err != nil {
				crp.Err = fmt.Errorf("GetFreePort for %s error %s", plugin.Name, err)
				rps = append(rps, &crp)
				continue
			}
			crp.port = port
		}
		args := []string{
			"-port", fmt.Sprint(crp.port),
			"-name", plugin.Name,
			"-type", plugin.Type,
		}
		for _, addArg := range plugin.AddArgs {
			args = append(args, addArg)
		}
		if tab != nil {
			pArgs, cOpts, err := tab(config.PluginsOptions)
			if err != nil {
				crp.Err = fmt.Errorf("PluginsOptions for %s error %s", plugin.Name, err)
				rps = append(rps, &crp)
				continue
			}
			for _, addArg := range pArgs {
				args = append(args, addArg)
			}
			crp.TlsClientOptions = cOpts
		}
		pExe, err := getExecutablePath(plugin)
		if err != nil {
			crp.Err = fmt.Errorf("plugin %s start error %s", plugin.Name, err)
			rps = append(rps, &crp)
			continue
		}
		plugin.ExecutablePath = pExe
		cmd := exec.Command(plugin.ExecutablePath, args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Start()
		crp.cmd = cmd
		if err != nil {
			crp.Err = fmt.Errorf("child %s start error %s", cmd, err)
		}
		rps = append(rps, &crp)
	}
	applyIfOK(rps, checkReadiness)
	return rps, nil
}

func RunConfFile(lgr *slog.Logger, confPath string) ([]*RunningPlugin, error) {
	bs, err := common.LoadFile(confPath)
	if err != nil {
		return nil, err
	}
	return RunConfData(lgr, string(bs), nil)
}

func Shutdown(rps []*RunningPlugin) {
	applyIfOK(rps, shutdown)
}

func WaitFor(rps []*RunningPlugin) {
	applyIfOK(rps, waitFor)
}

func Errors(rps []*RunningPlugin) []error {
	errs := []error{}
	for _, rp := range rps {
		if rp.Err != nil {
			errs = append(errs, rp.Err)
		}
	}
	return errs
}

func PluginFor(pName string, rps []*RunningPlugin) *RunningPlugin {
	for _, rp := range rps {
		if rp.Plugin.Name == pName {
			return rp
		}
	}
	return nil
}
