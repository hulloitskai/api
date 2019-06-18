// Package zero is a utility package that holds empty data types and no-op
// objects.
package zero

// Struct holds no information; it has a size of zero.
type Struct struct{}

// Empty is an instance of the zero Struct.
var Empty = Struct{}

// Interface says nothing; all objects implement the empty interface.
type Interface interface{}
