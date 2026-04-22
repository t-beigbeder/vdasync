package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/plugin"
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
		cmds, errs := plugin.RunConfFile(tf)
		if len(cmds) == 0 && len(errs) > 0 {
			common.Fatal(log, fmt.Errorf("RunPlugins failed %s", errs))
		}
		werrs := plugin.WaitFor(cmds)
		errs = append(errs, werrs...)
		if len(errs) > 0 {
			common.Fatal(log, fmt.Errorf("some child(ren) error(s) %s", errs))
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
	} else {
		common.Fatal(log, fmt.Errorf("neither root or plugin"))
	}
}
