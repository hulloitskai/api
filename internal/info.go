package internal

// Version is the current module version, and is set at compile-time
// with the following linker flag:
//   -X go.stevenxie.me/api/v2/internal.Version=$(VERSION)
var Version = "unset"

const (
	// Namespace is the module name, used for things like envvar prefixes.
	Namespace = "api"

	// ConfDir is the configuration directory corresponding to this module.
	ConfDir = "/etc/" + Namespace
)
