package task

import (
	"errors"
	"sentinels/global"
	"sentinels/model"
	"sentinels/store"
	"sync"
)

func buildBinder(device *model.Device) (*PointBinder, error) {
	//查询所有点位
	points := store.DbClient.SelectPointsByDeviceId(device.Id)
	if points == nil || len(points) == 0 {
		return nil, errors.New("no points found")
	}
	pb := &PointBinder{}
	if device.ProtocolType == global.ModbusTCP {
		//modbus
		collects, _ := store.DbClient.SelectCollectByDeviceId(device.Id)
		mc := &ModbusConvert{fcGroup: make(map[byte]map[uint16][]*model.Point)}
		mc = mc.convert(points).collect(collects).scatter()
		pb.loadModesPoints(mc)
	} else {
		//其他规约
	}
	return pb, nil
}

// PointBinder 点位集束器
type PointBinder struct {
	pss   []model.PointSnap
	lock  sync.Mutex
	index int
}

func (b *PointBinder) Next() model.PointSnap {
	b.lock.Lock()
	defer b.lock.Unlock()
	defer func() {
		b.index++
		if b.index >= len(b.pss) {
			b.index = 0
		}
	}()
	return b.pss[b.index]
}

func (b *PointBinder) loadModesPoints(convert *ModbusConvert) {
	for _, group := range convert.groupByPriority() {
		mps := &model.ModbusPointSnap{
			FuncCode:     group.funcCode,
			Points:       group.points,
			StartAddress: group.startAddress,
			EndAddress:   group.endAddress,
			Size:         group.size,
		}
		b.pss = append(b.pss, mps)
	}
}
