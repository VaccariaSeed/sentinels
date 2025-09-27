package main

import (
	"fmt"
	"os"
	"os/signal"
	_ "sentinels/api"
	"sentinels/global"
	"sentinels/model"
	"sentinels/task"
	_ "sentinels/task"
	"syscall"
	"time"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	// 通知signal包将SIGINT和SIGTERM信号转发到sigChan
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	global.SystemLog.Warnf("system stop by signal: %v", sig)
}

func init() {
	go func() {
		value := make(map[string]string)
		value["startAddr"] = "0x01"
		value["length"] = "0x02"
		time.Sleep(5 * time.Second)
		opt := &model.Operate{
			UniqueIdentifier: "2121212121",
			ReplySize:        0,
			SignType:         "address",
			Sign:             "test_coils",
			SendTime:         0,
			ValidityPeriod:   0,
			Cmd: &model.OperateCmd{
				Timeout:  0,
				CmdType:  "copyRead",
				FuncCode: "0x01",
				Value:    value,
			},
		}
		exec, err := task.GTP.Exec(opt)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("exec result:", exec)
	}()
}
