package cli

import (
	"errors"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"google.golang.org/grpc"
)

func GetServerOrPluginTls(cf *CommonFlagsType) (grpc.ServerOption, error) {
	if *cf.NoTlsFlag {
		return nil, nil
	}
	if *cf.CertFlag == "" {
		return nil, errors.New("GetServerOrPluginTls: no cert file")
	}
	if *cf.KeyFlag == "" {
		return nil, errors.New("GetServerOrPluginTls: no key file")
	}
	if *cf.ClientCaCertFlag == "" {
		return remote.GetSimpleTlsSOpt(*cf.CertFlag, *cf.KeyFlag)
	}
	return remote.GetMutualTlsSOpt(*cf.ClientCaCertFlag, *cf.CertFlag, *cf.KeyFlag)
}

func confStringMerge(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	if s2 == "" {
		return s1
	}
	return s1
}

func getClientMTls(caf, ccf, ckf string) (grpc.DialOption, error) {
	if ccf == "" {
		return nil, errors.New("getClientMTls: no client cert file")
	}
	if ckf == "" {
		return nil, errors.New("getClientMTls: no client key file")
	}
	return remote.GetMutualTlsCOpt(caf, ccf, ckf)
}


func GetClientServerTls(cf *CommonFlagsType, cfg *config.DataStoreType) (grpc.DialOption, error) {
	if cfg == nil {
		cfg = &config.DataStoreType{}
	}
	if *cf.NoTlsFlag || cfg.NoTls {
		return nil, nil
	}
	if *cf.TlsInsecFlag || cfg.Insecure {
		return remote.GetInsecureSkipVerifyCopt(), nil
	}
	caf := confStringMerge(*cf.CaCertFlag, cfg.CaCertPath)
	ccf := confStringMerge(*cf.ClientCertFlag, cfg.ClientCertPath)
	ckf := confStringMerge(*cf.ClientKeyFlag, cfg.ClientKeyPath)
	if ccf == "" {
		return nil, nil
	}
	if ckf == "" {
		return nil, errors.New("GetClientServerTls: no CA cert file")
	}
	return getClientMTls(caf, ccf, ckf)
}

func GetClientPluginTls(cf *CommonFlagsType, cfg *config.PluginsOptionsType) (grpc.DialOption, error) {
	if *cf.NoTlsPluginFlag || cfg.NoTls {
		return nil, nil
	}
	if *cf.TlsInsecPluginFlag || cfg.Insecure {
		return remote.GetInsecureSkipVerifyCopt(), nil
	}
	caf := confStringMerge(*cf.ClientCaCertFlag, cfg.CaCertPath)
	ccf := confStringMerge(*cf.ClientCertFlag, cfg.ClientCertPath)
	ckf := confStringMerge(*cf.ClientKeyFlag, cfg.ClientKeyPath)
	if ckf == "" {
		return nil, errors.New("GetClientPluginTls: no client CA cert file")
	}
	return getClientMTls(caf, ccf, ckf)
}

func GetPluginTlsOpts(cf *CommonFlagsType, cfg *config.PluginsOptionsType) (tlsArgs []string) {
	tlsArgs = []string{}
	if *cf.NoTlsPluginFlag || cfg.NoTls {
		tlsArgs = append(tlsArgs, "-notls")
		return
	}
	caf := confStringMerge(*cf.ClientCaCertFlag, cfg.CaCertPath)
	if caf != "" {
		tlsArgs = append(tlsArgs, "-clientca", caf)
	}
	ccf := confStringMerge(*cf.CertFlag, cfg.CertPath)
	if ccf != "" {
		tlsArgs = append(tlsArgs, "-cert", ccf)
	}
	ckf := confStringMerge(*cf.KeyFlag, cfg.KeyPath)
	if ccf != "" {
		tlsArgs = append(tlsArgs, "-key", ckf)
	}
	return
}
