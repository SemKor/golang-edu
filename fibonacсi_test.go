package main

import "testing"

func TestFibonacci(t *testing.T) {
	result := Fibonacci(10)
	expected := 55

	if result != expected {
		t.Errorf("expected %d got %d", expected, result)
	}
}
