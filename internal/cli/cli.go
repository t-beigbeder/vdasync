package cli

import "flag"

type CommonFlagsType struct {
	ConcurrencyFlag *int
	LogLevelFlag  *string
	LogFlag *string
}

func CommonFlags() *CommonFlagsType {
	return &CommonFlagsType{
		ConcurrencyFlag: flag.Int("conc", 0, "number of concurrent activities"),
		LogLevelFlag: flag.String("level", "", "log level, defaults to ERROR"),
		LogFlag: flag.String("log", "",
			"log file, defaults to dssacli-<pid>.log in temp dir, \"stderr\" is a known keyword"),
	}
}
