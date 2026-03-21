package main

import (
	"log"
	"os"

	"task4/cmd"
	"task4/internal/config"
	"task4/internal/logger"
)

var modeMap = map[string]func(cfg *config.Config){
	"":       cmd.RunClient,  
	"client": cmd.RunClient,
	"server": cmd.RunServer,
}

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Please provide path to config YAML")
	}

	cfgPath := os.Args[1] 

	
	cfg, err := config.Init(cfgPath)
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	logger.Init(cfg)

	modeFn, ok := modeMap[cfg.Mode]
	if !ok {
		log.Fatalf("unknown mode: %s", cfg.Mode)
	}

	modeFn(cfg)
}
