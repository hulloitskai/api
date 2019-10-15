package name

import "fmt"

// StructMethod names a struct method using the format:
//   (receiver).method
func StructMethod(receiver, method string) string {
	return fmt.Sprintf("(%s).%s", receiver, method)
}
