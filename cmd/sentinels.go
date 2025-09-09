package main

import (
	"os"
	"os/signal"
	_ "sentinels/api"
	"sentinels/global"
	"syscall"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	// 通知signal包将SIGINT和SIGTERM信号转发到sigChan
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	global.SystemLog.Warnf("system stop by signal: %v", sig)
}
