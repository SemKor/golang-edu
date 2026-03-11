package main

import (
	"fmt"
	"io"
	"os"
)

func foo(n int, w io.Writer) {
	for i := 0; i < n; i++ {
		fmt.Fprint(w, "Foo")
	}
}

func bar(n int, w io.Writer) {
	for i := 0; i < n; i++ {
		fmt.Fprint(w, "Bar")
	}
}

func FooBar(w io.Writer) {
	n := 10
	go foo(n, w)
	go bar(n, w)
}

func main() {
	FooBar(os.Stdout)
}
