package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

func foo(n int, w io.Writer, wg *sync.WaitGroup, fooCh chan struct{}, barCh  chan struct{}) {
	defer wg.Done()
	for i := 0; i < n; i++ {
		<- fooCh
		fmt.Fprint(w, "Foo")
		barCh <- struct{}{}
	}
}

func bar(n int, w io.Writer, wg *sync.WaitGroup, fooCh chan struct{}, barCh  chan struct{}) {
	defer wg.Done()
	for i := 0; i < n; i++ {
		<- barCh  
		fmt.Fprint(w, "Bar")
		fooCh <- struct{}{}
	}
}

func FooBar(w io.Writer, wg *sync.WaitGroup) {
	wg.Add(2)
	
	fooCh := make(chan struct{}, 1)
	barCh := make(chan struct{}, 1)
	
	fooCh <- struct{}{}
	n := 10
	go foo(n, w, wg, fooCh, barCh)
	go bar(n, w, wg, fooCh, barCh)
	
}

func main() {
	wg := sync.WaitGroup{}
	FooBar(os.Stdout, &wg)
	wg.Wait()
}
