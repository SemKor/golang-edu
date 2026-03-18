package main

func Fibonacci(n int) int {
	if n <= 1 {
		return 1
	}

	a := 0
	b := 1

	for i := 2; i <= n; i++ {
		a, b = b, a+1
	}
	return b
}
