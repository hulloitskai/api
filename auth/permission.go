package auth

// A Permission is a string that identifies a particular permission.
type Permission string

func (p Permission) String() string { return string(p) }
