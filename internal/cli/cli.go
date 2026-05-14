package cli

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/plugin"
	"google.golang.org/grpc"
)

type CommonFlagsType struct {
	ConfigFlag         *string
	ConcurrencyFlag    *int
	LogLevelFlag       *string
	LogFlag            *string
	SilentFlag         *bool
	VerboseFlag        *bool
	NoTlsFlag          *bool
	NoTlsPluginFlag    *bool
	TlsInsecFlag       *bool
	TlsInsecPluginFlag *bool
	ClientCaCertFlag   *string
	ClientCertFlag     *string
	ClientKeyFlag      *string
	CaCertFlag         *string
	CertFlag           *string
	KeyFlag            *string
}

func CommonFlags() *CommonFlagsType {
	return &CommonFlagsType{
		ConfigFlag:      flag.String("config", "", "configuration file, see documentation"),
		ConcurrencyFlag: flag.Int("conc", 0, "number of concurrent activities"),
		LogLevelFlag:    flag.String("level", "", "log level, defaults to ERROR"),
		LogFlag: flag.String("log", "",
			"log file, defaults to vdasync-<pid>.log in temp dir, \"stderr\" is a known keyword"),
		SilentFlag:         flag.Bool("silent", false, "no output"),
		VerboseFlag:        flag.Bool("verbose", false, "detailed output"),
		NoTlsFlag:          flag.Bool("notls", false, "insecure communication with servers over http"),
		NoTlsPluginFlag:    flag.Bool("notlsplugin", false, "insecure communication with plugins over http"),
		TlsInsecFlag:       flag.Bool("insec", false, "don't check certificate when communicating with server"),
		TlsInsecPluginFlag: flag.Bool("insecplugin", false, "don't check certificate when communicating with plugins"),
		ClientCaCertFlag:   flag.String("clientca", "", "client TLS certificate CA"),
		ClientCertFlag:     flag.String("clientcert", "", "client TLS certificate"),
		ClientKeyFlag:      flag.String("clientkey", "", "client TLS certificate key"),
		CaCertFlag:         flag.String("ca", "", "server or plugin TLS certificate CA"),
		CertFlag:           flag.String("cert", "", "server or plugin TLS certificate"),
		KeyFlag:            flag.String("key", "", "server or plugin TLS certificate key"),
	}
}

func parseUrlPlugin(protocol string) (pluginName string, err error) {
	pElts := strings.Split(protocol, "+")
	if len(pElts) > 1 {
		if pElts[1] != "dss" {
			err = fmt.Errorf("unknown protocol %s", pElts[1])
			return
		}
		pluginName = pElts[0]
		return
	}
	if pElts[0] != "dss" {
		err = fmt.Errorf("unknown protocol %s", pElts[0])
		return
	}
	return
}

func ParseUrl(url string) (pluginName, host string, port int, rootPath string, err error) {
	// relativeRootPath
	// /rootPath
	// [pluginName+]dss:/rootPath
	// [pluginName+]dss://host[:port]/rootPath
	urlElts := strings.Split(url, "://")
	if len(urlElts) > 1 {
		if pluginName, err = parseUrlPlugin(urlElts[0]); err != nil {
			return
		}
		rscElts := strings.Split(urlElts[1], "/")
		if len(rscElts) <= 1 {
			err = fmt.Errorf("resource part %s has no host name", urlElts[1])
			return
		}
		// [pluginName+]dss://host[:port]/rootPath
		hp := strings.Split(rscElts[0], ":")
		if len(hp) > 1 {
			var port64 int64
			if port64, err = strconv.ParseInt(hp[1], 10, 0); err != nil {
				return
			}
			port = int(port64)
		}
		host = hp[0]
		rootPath = "/" + strings.Join(rscElts[1:], "/")
	} else {
		urlElts := strings.Split(url, ":")
		if len(urlElts) > 1 {
			if pluginName, err = parseUrlPlugin(urlElts[0]); err != nil {
				return
			}
			// [pluginName+]dss:/rootPath
			rootPath = urlElts[1]
		} else {
			// relativeRootPath
			// /rootPath
			rootPath = urlElts[0]
		}
	}
	rootPath, err = NormalizeRoot(rootPath)
	return
}

func NormalizeRoot(rootPath string) (string, error) {
	return filepath.Abs(rootPath)
}

func RunPlugins(confData string, cf *CommonFlagsType) ([]*plugin.RunningPlugin, error) {
	tab := func(cfg *config.PluginsOptionsType) ([]string, grpc.DialOption, error) {
		dop, err := GetClientPluginTls(cf, cfg)
		return GetPluginTlsOpts(cf, cfg), dop, err
	}
	rps, err := plugin.RunConfData(confData, tab)
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
