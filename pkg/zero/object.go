// Package zero is a utility package that holds empty data types and no-op
// objects.
package zero

// Struct holds no information; it has a size of zero.
type Struct = struct{}

// Empty returns an empty struct.
func Empty() Struct { return struct{}{} }

// Interface says nothing; all objects implement the empty interface.
type Interface = interface{}
