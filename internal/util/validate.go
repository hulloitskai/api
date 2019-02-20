package util

import (
	"fmt"

	"golang.org/x/xerrors"
	vld "gopkg.in/validator.v2"
)

// NewValidate builds a Validate that uses the 'valid' struct tag, and supports
// custom validation for any type that implements the Validator interface.
func NewValidate() *vld.Validator {
	valid := vld.NewValidator()
	valid.SetTag("valid")
	if err := valid.SetValidationFunc(
		"custom",
		func(v interface{}, param string) error {
			validator, ok := v.(Validator)
			if !ok {
				return fmt.Errorf("%s does not implement util.Validator", param)
			}
			return validator.Validate()
		},
	); err != nil {
		panic(xerrors.Errorf("util: configuring custom validation: %w", err))
	}
	return valid
}

// A Validator can validate itself.
type Validator interface {
	Validate() error
}

// Validate performs a validation using the default validator.Validate.
func Validate(v interface{}) error {
	if defaultValidator == nil {
		defaultValidator = NewValidate()
	}
	return defaultValidator.Validate(v)
}

var defaultValidator *vld.Validator
