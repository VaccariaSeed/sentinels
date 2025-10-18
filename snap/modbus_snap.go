package snap

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sentinels/global"
	"sentinels/model"
)

type ModbusPointSnap struct {
	FuncCode     byte
	Points       map[uint16][]*model.Point //多个点位时，这里面是连续的地址，单个点位时，这里面只有一个地址
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

func (m *ModbusPointSnap) Point(key interface{}) ([]*model.Point, error) {
	if u, ok := key.(uint16); ok {
		return m.Points[u], nil
	}
	return nil, errors.New("invalid point key")
}

func (m *ModbusPointSnap) Parse(resp []byte) (map[string]interface{}, error) {
	if resp == nil || len(resp) < 1 {
		return nil, errors.New("invalid resp, it is empty")
	}
	result := make(map[string]interface{})
	for index := m.StartAddress; index <= m.EndAddress; index++ {
		points := m.Points[index]
		var value interface{}
		var err error
		for _, p := range points {
			//获取最基本的值
			switch p.DataType {
			case global.DTBit:
				value, err = m.bitFlush(resp, index, m.StartAddress)
			case global.DTInt8:
				value, err = m.multipleFlushInt8(resp, index, m.StartAddress, p)
			case global.DTInt16:
				value, err = m.multipleFlushInt16(resp, index, m.StartAddress, p)
			case global.DTInt32:
				value, err = m.multipleFlushInt32(resp, index, m.StartAddress, p)
			case global.DTInt64:
				value, err = m.multipleFlushInt64(resp, index, m.StartAddress, p)
			case global.DTByte:
				value, err = m.multipleFlushByte(resp, index, m.StartAddress, p)
			case global.DTUint16:
				value, err = m.multipleFlushUint16(resp, index, m.StartAddress, p)
			case global.DTUint32:
				value, err = m.multipleFlushUint32(resp, index, m.StartAddress, p)
			case global.DTUint64:
				value, err = m.multipleFlushUint64(resp, index, m.StartAddress, p)
			case global.DTFloat32:
				value, err = m.multipleFlushFloat32(resp, index, m.StartAddress, p)
			case global.DTFloat64:
				value, err = m.multipleFlushFloat64(resp, index, m.StartAddress, p)
			default:
				return nil, errors.New("invalid point data type")
			}
			if err != nil {
				return nil, err
			}
			result[p.Tag] = value
		}
	}
	return result, nil
}

func (m *ModbusPointSnap) bitFlush(resp []byte, index uint16, address uint16) (int8, error) {
	var binaryStr string
	for i := len(resp) - 1; i >= 0; i-- {
		binaryStr += fmt.Sprintf("%08b", resp[i])
	}
	if binaryStr == "" {
		return 0, errors.New("invalid point size")
	}
	//翻转数组
	runes := []rune(binaryStr)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	binaryStr = string(runes)
	i := index - address
	return int8(binaryStr[i]), nil
}

func (m *ModbusPointSnap) multipleFlush(resp []byte, index, address uint16, endianness, bc string, size, startBit, endBit int) ([]byte, error) {
	if startBit < 0 || endBit < 0 {
		return nil, errors.New("invalid start bit or end bit")
	}
	in := int(index - address)
	if in < 0 || in >= len(resp) {
		return nil, errors.New("start index invalid point size")
	}
	en := in + size
	if en > len(resp) {
		return nil, errors.New("end index invalid point size")
	}
	values := resp[in:en]
	//大端小端
	if endianness == global.LittleEndian {
		//小端
		for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
			values[i], values[j] = values[j], values[i]
		}
	}
	if bc == global.SingleBit {
		//单bit
		if startBit >= size*8 {
			return nil, fmt.Errorf("start bit is too large (%d)", startBit)
		}
		byteIndex := startBit / 8
		bitOffset := 7 - (startBit % 8)
		return append(make([]byte, size-1), byte((values[byteIndex]>>bitOffset)&1)), nil
	} else if bc == global.MultipleBit {
		if startBit <= endBit {
			return nil, errors.New("start bit <= end bit")
		}
		if endBit >= size*8 {
			return nil, fmt.Errorf("end bit is too large (%d)", startBit)
		}
		result := make([]byte, len(values))
		// 计算提取的比特数和需要左移的位数
		bitCount := endBit - startBit + 1
		leftShift := size - bitCount
		for i := 0; i < bitCount; i++ {
			srcBitPos := startBit + i
			dstBitPos := leftShift + i

			// 计算源字节和位偏移（大端序）
			srcByteIndex := srcBitPos / 8
			srcBitOffset := 7 - (srcBitPos % 8)
			srcBit := (values[srcByteIndex] >> srcBitOffset) & 1

			// 计算目标字节和位偏移（大端序）
			dstByteIndex := dstBitPos / 8
			dstBitOffset := 7 - (dstBitPos % 8)

			// 设置目标位
			if srcBit == 1 {
				result[dstByteIndex] |= 1 << dstBitOffset
			}
		}
		return result, nil
	} else {
		return values, nil
	}
}

func (m *ModbusPointSnap) execNumber(luaTemp string, value, multiplier, offset float64) (float64, error) {
	//lua
	result, err := sl.execNumber(luaTemp, value)
	if err != nil {
		return 0, err
	}
	//倍率
	if multiplier != 0 {
		result = result * multiplier
	}
	//偏移量
	return result - offset, nil
}

func (m *ModbusPointSnap) multipleFlushInt8(resp []byte, index uint16, address uint16, p *model.Point) (float64, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 1, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := float64(int8(values[0]))
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}

func (m *ModbusPointSnap) multipleFlushInt16(resp []byte, index uint16, address uint16, p *model.Point) (interface{}, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 2, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := float64(int16(binary.BigEndian.Uint16(values)))
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}

func (m *ModbusPointSnap) multipleFlushInt32(resp []byte, index uint16, address uint16, p *model.Point) (interface{}, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 4, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := float64(int32(binary.BigEndian.Uint32(values)))
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}

func (m *ModbusPointSnap) multipleFlushInt64(resp []byte, index uint16, address uint16, p *model.Point) (interface{}, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 8, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := float64(int64(binary.BigEndian.Uint64(values)))
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}

func (m *ModbusPointSnap) multipleFlushByte(resp []byte, index uint16, address uint16, p *model.Point) (interface{}, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 1, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := float64(values[0])
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}

func (m *ModbusPointSnap) multipleFlushUint16(resp []byte, index uint16, address uint16, p *model.Point) (interface{}, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 2, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := float64(binary.BigEndian.Uint16(values))
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}

func (m *ModbusPointSnap) multipleFlushUint32(resp []byte, index uint16, address uint16, p *model.Point) (interface{}, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 4, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := float64(binary.BigEndian.Uint32(values))
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}

func (m *ModbusPointSnap) multipleFlushUint64(resp []byte, index uint16, address uint16, p *model.Point) (interface{}, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 8, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := float64(binary.BigEndian.Uint64(values))
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}

func (m *ModbusPointSnap) multipleFlushFloat32(resp []byte, index uint16, address uint16, p *model.Point) (interface{}, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 4, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := float64(math.Float32frombits(binary.BigEndian.Uint32(values)))
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}

func (m *ModbusPointSnap) multipleFlushFloat64(resp []byte, index uint16, address uint16, p *model.Point) (interface{}, error) {
	values, err := m.multipleFlush(resp, index, address, p.Endianness, p.BitCalculation, 8, p.StartBit, p.EndBit)
	if err != nil {
		return 0, err
	}
	result := math.Float64frombits(binary.BigEndian.Uint64(values))
	return m.execNumber(p.LuaExpression, result, p.Multiplier, p.Offset)
}
