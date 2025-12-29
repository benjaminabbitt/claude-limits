package version

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X ClaudeLimits/internal/version.Version=$(versionator)"
var Version = "dev"
