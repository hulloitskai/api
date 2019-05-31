package info

// Version is the program version, set during compile time using:
//   -ldflags -X github.com/stevenxie/api/internal.Version=$(VERSION)
var Version = "unset"

// Namespace is the project namespace, to be used as prefixes for environment
// variables, etc.
const Namespace = "api"
