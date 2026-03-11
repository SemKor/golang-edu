package main

import (
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
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}
	logger.Init(cfg)
	modeFn, ok := modeMap[cfg.Mode]
	if !ok {
		panic("unknown mode!")
	}
	modeFn(cfg)
}
