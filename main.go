package main

import (
	"task5/cmd"
	"task5/internal/config"
)

func main() {
	cfg := config.Init()
	ctlFunc, ok := cmd.ModeMap[cfg.Mode]
	if !ok {
		panic("incorrect program run mode")
	}
	ctl := ctlFunc(cfg)
	ctl.Wait()
}
