package config

// Version can be set at link time
// -ldflags "-X config.Version=$(git describe --tags)"
var Version string

func GetVersion() string {
	if Version != "" {
		return Version
	}
	return "dev"
}
