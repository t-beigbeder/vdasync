package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/dssagrpc"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/otvl_dtacsy/internal/plugin"
	"github.com/t-beigbeder/otvl_dtacsy/internal/remote"
)

const CliYamlConfigTest string = `
plugins:
- name: localFilesSample
  type: localFiles
  port: 10314
  # executablePath: /home/dv-user/locgit/otvl_dtacsy/bin/test
  addArgs:
  - "-is-plugin"
- name: noTypeErrorSample
  addArgs:
  - "-is-plugin"
`

func main() {
	var (
		isRoot      bool
		isPlugin    bool
		isFatal     bool
		port        int
		name, type_ string
	)
	flag.BoolVar(&isRoot, "is-root", false, "")
	flag.BoolVar(&isPlugin, "is-plugin", false, "")
	flag.BoolVar(&isFatal, "is-fatal", false, "")
	flag.IntVar(&port, "port", 0, "grpc server port (plugin only)")
	flag.StringVar(&name, "name", "", "plugin name")
	flag.StringVar(&type_, "type", "", "plugin type")
	flag.Parse()
	log := common.GetLogger()

	if isRoot {
		log = log.With("app", "main")
		tf := path.Join(os.TempDir(), "testgrpc.yml")
		err := common.WriteFile(tf, []byte(CliYamlConfigTest))
		if err != nil {
			common.Fatal(log, err)
		}
		rps, err := plugin.RunConfFile(tf)
		if err != nil {
			common.Fatal(log, fmt.Errorf("RunConfFile failed %s", err))
		}
		if len(plugin.Errors(rps)) != 0 {
			log.Warn("RunConfFile raised some errors", "errors", plugin.Errors(rps))
		}
		var rp *plugin.RunningPlugin
		for _, rpc := range rps {
			if rpc.Err == nil {
				rp = rpc
				break
			}
		}
		if rp != nil {
			des, err := rp.Client.List(context.Background(), &dssagrpc.Path{Path: []string{"."}})
			if err == nil {
				for _, en := range des.Entries {
					log.Debug("List result", "plugin", rp.Plugin.Name, "entry", en.Path)
				}
			} else {
				log.Warn("List error", "plugin", rp.Plugin.Name, "error", err)
			}
		}
		plugin.Shutdown(rps)
		if len(plugin.Errors(rps)) != 0 {
			log.Warn("RunConfFile and/or Shutdown raised some errors", "errors", plugin.Errors(rps))
		}
		plugin.WaitFor(rps)
		if len(plugin.Errors(rps)) != 0 {
			log.Warn("RunConfFile/Shutdown and/or WaitFor raised some errors", "errors", plugin.Errors(rps))
		}
	} else if isPlugin {
		log = log.With("app", "child")
		log.Info("started", "args", fmt.Sprint(os.Args))
		if type_ == "" {
			common.Fatal(log, fmt.Errorf("plugin requires a type"))
		}
		if isFatal {
			common.Fatal(log, fmt.Errorf("child failing on demand"))
		}
		done := make(chan bool)
		cb := func() {
			log.Debug("shutdownCb called, closing done")
			close(done)
		}
		_, _, err := remote.RunOpeDssaServer(context.Background(), "localhost", port, nil, localfiles.MakeLocalFilesDssa(), cb)
		<-done
		if err != nil {
			common.Fatal(log, fmt.Errorf("RunOpeDssaServer failed %s", err))
		}
	} else {
		common.Fatal(log, fmt.Errorf("neither root or plugin"))
	}
	log.Info("stopped")
}
