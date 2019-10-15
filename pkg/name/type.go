package name

import (
	"reflect"

	"go.stevenxie.me/gopkg/zero"
)

// OfTypeFull returns the full type name of a value.
func OfTypeFull(v zero.Interface) string {
	t := getType(v)
	return t.PkgPath() + "." + t.Name()
}

// OfType returns the
func OfType(v zero.Interface) string {
	t := getType(v)
	return t.String()
}

func getType(v zero.Interface) reflect.Type {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Ptr, reflect.Interface:
		t = t.Elem()
	}
	return t
}
