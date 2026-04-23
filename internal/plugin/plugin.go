package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/remote"
	"github.com/t-beigbeder/otvl_dtacsy/opegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RunningPlugin struct {
	config *config.CliConfig
	plugin *config.PluginType
	port   int
	cmd    *exec.Cmd
	client remote.OpeDssaClient
	err    error
}

func checkReadiness(rp *RunningPlugin) {
	if rp.config.PluginReadyRetries == 0 || rp.config.PluginReadyTimeout == "" {
		return
	}
	retryDelay, err := time.ParseDuration(rp.config.PluginReadyTimeout)
	if err != nil {
		rp.err = fmt.Errorf("ParseDuration failed with error %s", err)
		return
	}
	opts := []grpc.DialOption{}
	switch rp.config.PluginTransportCredentials {
	case "insecure":
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	default:
		rp.err = fmt.Errorf("incorrect PluginTransportCredentials: %s", rp.config.PluginTransportCredentials)
		return
	}
	for count := 0; count < rp.config.PluginReadyRetries; count++ {
		cli, err := remote.CheckServerReadiness(fmt.Sprintf("%s:%d", rp.config.PluginAddress, rp.port), opts...)
		rp.err = err
		if err == nil {
			rp.client = cli
			break
		}
		time.Sleep(time.Duration(retryDelay))
		retryDelay *= 2
	}
}

func shutdown(rp *RunningPlugin) {
	_, rp.err = rp.client.Shutdown(context.Background(), &opegrpc.Value{Value: "100ms"})
}

func waitFor(rp *RunningPlugin) {
	err := rp.cmd.Wait()
	if err != nil {
		rp.err = fmt.Errorf("child %s wait error %s", rp.cmd, err)
	}
}

func applyIfOK(rps []*RunningPlugin, run func(*RunningPlugin)) {
	for _, rp := range rps {
		if rp.cmd == nil || rp.err != nil {
			continue
		}
		run(rp)
	}
}

func RunConfFile(confPath string) ([]*RunningPlugin, error) {
	config, err := config.Load(confPath)
	if err != nil {
		return nil, err
	}
	rps := []*RunningPlugin{}
	for _, plugin := range config.Plugins {
		crp := RunningPlugin{config: config, plugin: plugin, port: plugin.Port}
		if crp.port == 0 {
			port, err := common.GetFreePort()
			if err != nil {
				crp.err = fmt.Errorf("GetFreePort for %s error %s", plugin.Name, err)
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
		cmd := exec.Command(plugin.ExecutablePath, args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Start()
		crp.cmd = cmd
		if err != nil {
			crp.err = fmt.Errorf("child %s start error %s", cmd, err)
		}
		rps = append(rps, &crp)
	}
	applyIfOK(rps, checkReadiness)
	return rps, nil
}

func RunConfData(tempFilePath string, conf string) ([]*RunningPlugin, error) {
	if err := common.WriteFile(tempFilePath, []byte(conf)); err != nil {
		return nil, err
	}
	return RunConfFile(tempFilePath)
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
		if rp.err != nil {
			errs = append(errs, rp.err)
		}
	}
	return errs
}
