package errors

// An WithCode is an error that can be identified by an integer code.
type WithCode interface {
	error
	Code() int
}
