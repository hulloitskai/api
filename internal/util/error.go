package util

import (
	errors "golang.org/x/xerrors"
)

type wrappedError struct{ Err, Cause error }

// Error implements errors.Error.
func (we *wrappedError) Error() string { return we.Err.Error() }

// Unwrap implements errors.Unwrapper.
func (we *wrappedError) Unwrap() error { return we.Cause }

// FormatError implements errors.Formatter.
func (we *wrappedError) FormatError(p errors.Printer) (next error) {
	if err, ok := we.Err.(errors.Formatter); ok {
		return err.FormatError(p)
	}
	return we.Err
}

// WrapCause wraps an error with its cause.
func WrapCause(err, cause error) error {
	return &wrappedError{err, cause}
}
