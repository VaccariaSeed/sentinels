package task

import (
	"fmt"
	"sentinels/global"
	"sentinels/model"
	"sentinels/snap"
	"sentinels/store"
	"sync"
)

func buildBinder(device *model.Device) (*PointBinder, error) {
	//查询所有点位
	points := store.DbClient.SelectPointsByDeviceId(device.Id)
	pb := &PointBinder{}
	if points == nil || len(points) == 0 {
		return pb, nil
	}
	switch device.ProtocolType {
	case global.ModbusTCP, global.ModbusRTU:
		//modbus
		collects, _ := store.DbClient.SelectCollectByDeviceId(device.Id)
		mc := &ModbusConvert{fcGroup: make(map[byte]map[uint16][]*model.Point)}
		mc = mc.convert(points).collect(collects).scatter()
		pb.loadModesPoints(mc)
	default:
		return nil, fmt.Errorf("unsupported protocol type: %s", device.ProtocolType)
	}

	return pb, nil
}

// PointBinder 点位集束器
type PointBinder struct {
	pss   []snap.PointSnap
	lock  sync.Mutex
	index int
}

func (b *PointBinder) Next() snap.PointSnap {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.pss == nil || len(b.pss) == 0 {
		return nil
	}
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
		if group == nil {
			continue
		}
		mps := &snap.ModbusPointSnap{
			FuncCode:     group.funcCode,
			Points:       group.points,
			StartAddress: group.startAddress,
			EndAddress:   group.endAddress,
			Size:         group.size,
		}
		b.pss = append(b.pss, mps)
	}
}
