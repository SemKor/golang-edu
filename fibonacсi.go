package main

func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}

	a := 0
	b := 1

	for i := 2; i <= n; i++ {
		a, b = b, a + b
	}
	
	return b
}
