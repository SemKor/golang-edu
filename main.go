package main

import (
	"flag"
	"fmt"
	"log"
	"task1/config"
)

func main() {

	n := flag.Int("n", -1, "fibonacci number")
	configPath := flag.String("c", "", "path to config file")

	flag.Parse()

	var number int

	if *n != -1 {
		number = *n
	} else if *configPath != "" {

		cfg, err := config.Init(*configPath)
		if err != nil {
			log.Fatal(err)
		}

		number = cfg.N
	} else {
		log.Fatal("provide -n or -c flag")
	}

	result := Fibonacci(number)

	fmt.Printf("Fibonacci(%d) = %d\n", number, result)
}
