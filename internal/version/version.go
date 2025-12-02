package version

// Version is set at build time via -ldflags "-X container-test-cli/internal/version.Version=vX.Y.Z".
var Version = "dev"
