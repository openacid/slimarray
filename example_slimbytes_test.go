package slimarray_test

import (
	"fmt"

	"github.com/openacid/slimarray"
)

func ExampleSlimBytes() {

	records := [][]byte{
		[]byte("SlimBytes"),
		[]byte("is"),
		[]byte("an"),
		[]byte("array"),
		[]byte("of"),
		[]byte("var-length"),
		[]byte("records(a"),
		[]byte("record"),
		[]byte("is"),
		[]byte("a"),
		[]byte("[]byte"),
		[]byte("which"),
		[]byte("is"),
		[]byte("indexed"),
		[]byte("by"),
		[]byte("SlimArray"),
	}

	a, err := slimarray.NewBytes(records)
	_ = err

	for i := 0; i < 16; i++ {
		fmt.Print(string(a.Get(int32(i))), " ")
	}
	fmt.Println()

	// Output:
	// SlimBytes is an array of var-length records(a record is a []byte which is indexed by SlimArray
}
