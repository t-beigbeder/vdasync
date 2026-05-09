package cli

import "flag"

type CommonFlagsType struct {
	ConcurrencyFlag  *int
	LogLevelFlag     *string
	LogFlag          *string
	SilentFlag       *bool
	VerboseFlag      *bool
	NoTlsFlag        *bool
	NoTlsFlagPlugin  *bool
	ClientCaCertFlag *string
	ClientCertFlag   *string
	ClientKeyFlag    *string
	CaCertFlag       *string
	CertFlag         *string
	KeyFlag          *string
}

func CommonFlags() *CommonFlagsType {
	return &CommonFlagsType{
		ConcurrencyFlag: flag.Int("conc", 0, "number of concurrent activities"),
		LogLevelFlag:    flag.String("level", "", "log level, defaults to ERROR"),
		LogFlag: flag.String("log", "",
			"log file, defaults to dssacli-<pid>.log in temp dir, \"stderr\" is a known keyword"),
		SilentFlag:       flag.Bool("silent", false, "no output"),
		VerboseFlag:      flag.Bool("verbose", false, "detailed output"),
		NoTlsFlag:        flag.Bool("notls", false, "unsecure communication with servers over http"),
		NoTlsFlagPlugin:  flag.Bool("notlsplugin", false, "unsecure communication with plugins over http"),
		ClientCaCertFlag: flag.String("clientca", "", "client TLS certificate CA"),
		ClientCertFlag:   flag.String("clientcert", "", "client TLS certificate"),
		ClientKeyFlag:    flag.String("clientkey", "", "client TLS certificate key"),
		CaCertFlag:       flag.String("ca", "", "server or plugin TLS certificate CA"),
		CertFlag:         flag.String("cert", "", "server or plugin TLS certificate"),
		KeyFlag:          flag.String("key", "", "server or plugin TLS certificate key"),
	}
}
