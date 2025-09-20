package task

import (
	"sentinels/model"
	"sort"
	"strconv"
)

type groupFunc struct {
	funcCode     byte
	points       map[uint16][]*model.Point
	startAddress uint16
	endAddress   uint16
	size         byte
	isFirst      bool
	priority     byte
}

func (f *groupFunc) appendPoints(addr uint16, points []*model.Point) {
	if !f.isFirst {
		f.startAddress = addr
		f.isFirst = true
	} else {
		if f.startAddress > addr {
			f.startAddress = addr
		}
	}
	if f.endAddress < addr {
		f.endAddress = addr
	}
	f.size++
	for _, point := range points {
		if point.Priority > f.priority {
			f.priority = point.Priority
		}
	}
	f.points[addr] = points
}

type ModbusConvert struct {
	fcGroup map[byte]map[uint16][]*model.Point //map[funcCode]map[address][]point
	gf      []*groupFunc
}

func (m *ModbusConvert) groupByPriority() []*groupFunc {
	sort.Slice(m.gf, func(i, j int) bool {
		return m.gf[i].priority > m.gf[j].priority
	})
	return m.gf
}

func (m *ModbusConvert) convert(points []*model.Point) *ModbusConvert {
	for _, point := range points {
		// 第一级：funcCode
		fc, _ := strconv.ParseUint(point.FunctionCode, 0, 8)
		if _, exists := m.fcGroup[byte(fc)]; !exists {
			m.fcGroup[byte(fc)] = make(map[uint16][]*model.Point)
		}
		// 第二级：address
		addr, _ := strconv.ParseUint(point.Address, 0, 16)
		m.fcGroup[byte(fc)][uint16(addr)] = append(m.fcGroup[byte(fc)][uint16(addr)], point)
	}
	return m
}

func (m *ModbusConvert) collect(collects []*model.Collect) *ModbusConvert {
	for _, collect := range collects {
		funcCode := collect.RuleFuncCode
		if _, ok := m.fcGroup[funcCode]; !ok {
			continue
		}
		//map[address][]*point
		points := m.fcGroup[funcCode]
		start, _ := strconv.ParseUint(collect.StartPoint, 0, 16)
		end, _ := strconv.ParseUint(collect.EndPoint, 0, 16)
		if start >= end {
			continue
		}
		gf := &groupFunc{funcCode: funcCode, points: map[uint16][]*model.Point{}}
		for index := start; index <= end; index++ {
			for addr, point := range points {
				if uint16(index) == addr {
					gf.appendPoints(addr, point)
					delete(points, addr)
				}
			}
		}
		m.gf = append(m.gf, gf)
	}
	return m
}

func (m *ModbusConvert) scatter() *ModbusConvert {
	//groupAddr map[address][]*point
	for funcCode, groupAddr := range m.fcGroup {
		if len(groupAddr) == 0 {
			continue
		}
		points := m.modbusContinuous(groupAddr)
		// pointGroup map[uint16][]*model.Point
		for _, pointGroup := range points {
			gf := &groupFunc{funcCode: funcCode, points: map[uint16][]*model.Point{}}
			for addr, point := range pointGroup {
				gf.appendPoints(addr, point)
			}
			m.gf = append(m.gf, gf)
		}
	}
	return m
}

func (m *ModbusConvert) modbusContinuous(addressMap map[uint16][]*model.Point) []map[uint16][]*model.Point {
	numericAddresses := make([]uint16, 0, len(addressMap))
	pointStore := make(map[uint16][]*model.Point)

	for addr, point := range addressMap {
		numericAddresses = append(numericAddresses, addr)
		pointStore[addr] = point
	}

	// 2. 按地址排序
	sort.Slice(numericAddresses, func(i, j int) bool {
		return numericAddresses[i] < numericAddresses[j]
	})

	// 3. 分组
	var groups []map[uint16][]*model.Point
	var currentGroup map[uint16][]*model.Point
	var prevAddr uint16

	for i, currentAddr := range numericAddresses {
		if i == 0 {
			// 第一个地址，开始新组
			currentGroup = make(map[uint16][]*model.Point)
			currentGroup[currentAddr] = pointStore[currentAddr]
			prevAddr = currentAddr
			continue
		}

		// 检查是否连续且不超过最大大小
		isContinuous := currentAddr == prevAddr+1
		withinLimit := len(currentGroup) < 125

		if isContinuous && withinLimit {
			currentGroup[currentAddr] = pointStore[currentAddr]
			prevAddr = currentAddr
		} else {
			// 当前组结束，开始新组
			groups = append(groups, currentGroup)
			currentGroup = make(map[uint16][]*model.Point)
			currentGroup[currentAddr] = pointStore[currentAddr]
			prevAddr = currentAddr
		}
	}

	// 添加最后一个组
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}
	return groups
}
