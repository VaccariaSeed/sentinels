package task

import (
	"fmt"
	"sentinels/command"
	"sentinels/global"
	"sentinels/store"
)

var GTP *GaTaskPool

type GaTaskPool struct {
	GTPSnapshotById        map[string]*GaTaskProcessor
	GTPSnapshotByTableFlag map[string]*GaTaskProcessor
}

func (g *GaTaskPool) Append(id string, tableFlag string, gtpr *GaTaskProcessor) {
	g.GTPSnapshotById[id] = gtpr
	g.GTPSnapshotByTableFlag[tableFlag] = gtpr
}

func init() {
	GTP = &GaTaskPool{
		GTPSnapshotById:        map[string]*GaTaskProcessor{},
		GTPSnapshotByTableFlag: map[string]*GaTaskProcessor{},
	}
	//查询切入的设备
	devices := store.DbClient.SelectCutInDevice()
	if len(devices) == 0 {
		return
	}
	for _, device := range devices {
		//创建调度器
		gtp, err := NewGaTaskProcessor(devices[0])
		if err != nil {
			global.SystemLog.Error(fmt.Sprintf("id:%s name:%s type:%s err:%s", device.Id, device.Name, device.Code, err.Error()))
			continue
		}
		go func() {
			_ = gtp.Start()
		}()
		GTP.Append(device.Id, device.Table, gtp)
	}
}

func (g *GaTaskPool) Exec(opt *command.ControlCarrier) ([]byte, error) {
	err := opt.Check()
	if err != nil {
		return nil, err
	}
	var gtp *GaTaskProcessor
	signType, sign := opt.ObtainSign()
	if signType == global.LogoTypeId {
		gtp = g.GTPSnapshotById[sign]
	} else {
		gtp = g.GTPSnapshotByTableFlag[sign]
	}
	if gtp == nil {
		return nil, fmt.Errorf("not found GaTaskProcessor by: %s, use:%s", sign, signType)
	}
	//开始执行命令
	var resp []byte
	for index := 0; index <= opt.ReplySize; index++ {
		resp, err = gtp.operate(opt.Cmd)
		if err != nil {
			continue
		}
		return resp, nil
	}
	return nil, err
}
