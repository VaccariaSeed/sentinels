package model

import (
	"encoding/json"
	"errors"
)

var _ PointSnap = (*ModbusPointSnap)(nil)

type PointSnap interface {
	Address() []byte //地址
	Length() byte    //数量
	FunctionCode() []byte
	String() string
	Point(key interface{}) ([]*Point, error)
	Parse(resp []byte) (map[string]interface{}, error)
}

type ModbusPointSnap struct {
	FuncCode     byte
	Points       map[uint16][]*Point //多个点位时，这里面是连续的地址，单个点位时，这里面只有一个地址
	StartAddress uint16
	EndAddress   uint16
	Size         byte //多少个点位
}

func (m *ModbusPointSnap) Address() []byte {
	return []byte{byte(m.StartAddress >> 8), byte(m.StartAddress)}
}

func (m *ModbusPointSnap) Length() byte {
	return m.Size
}

func (m *ModbusPointSnap) FunctionCode() []byte {
	return []byte{m.FuncCode}
}

func (m *ModbusPointSnap) String() string {
	data, _ := json.Marshal(m)
	return string(data)
}

func (m *ModbusPointSnap) Point(key interface{}) ([]*Point, error) {
	if u, ok := key.(uint16); ok {
		return m.Points[u], nil
	}
	return nil, errors.New("invalid point key")
}

func (m *ModbusPointSnap) Parse(resp []byte) (map[string]interface{}, error) {
	i := 0
	result := make(map[string]interface{})
	for index := m.StartAddress; index <= m.EndAddress; index++ {
		points := m.Points[index]
		if i >= len(resp) {
			return nil, errors.New("invalid point size")
		}
		for _, p := range points {
			value, err := p.BaseParse(resp[i], resp[i+1])
			if err != nil {
				return nil, err
			}
			result[p.Tag] = value
		}
		i += 2
	}
	return result, nil
}
