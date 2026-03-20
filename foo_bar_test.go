package main

import (
	"bytes"
	"sync"
	"testing"
)

func TestFooBar(t *testing.T) {
	var output bytes.Buffer
	wg := sync.WaitGroup{}
	want := "FooBarFooBarFooBarFooBarFooBarFooBarFooBarFooBarFooBarFooBar"
	FooBar(&output, &wg)
	wg.Wait()
	if want != output.String() {
		t.Error("got value that mismatch expected - ", output.String())
	}
}


// Данный тест необходим, чтобы с большей вероятностью убедиться
// в том, что наше решение не зависит от того, как шедулер запускает выполнение горутин
func TestFooBarN(t *testing.T) {
	for i := 0; i < 1000; i++ {
		TestFooBar(t)
	}
}
