package name_test

import (
	"fmt"

	"go.stevenxie.me/api/pkg/name"
)

type SomeStruct struct{}
type SomeIface interface{}

func ExampleOfTypeFull() {
	fmt.Println(name.OfTypeFull((*SomeStruct)(nil)))
	fmt.Println(name.OfTypeFull((*SomeIface)(nil)))
	// Output:
	// go.stevenxie.me/api/pkg/name_test.SomeStruct
	// go.stevenxie.me/api/pkg/name_test.SomeIface
}
