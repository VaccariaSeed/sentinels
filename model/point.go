package model

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sentinels/global"
)

type dataTypeInfo struct {
	byteLen int
	maxBit  int
	signed  bool
	isArray bool
}

var typeInfoMap = map[string]dataTypeInfo{}

func init() {
	typeInfoMap = map[string]dataTypeInfo{
		global.DTInt8:   {1, 7, true, false},
		global.DTUint8:  {1, 7, false, false},
		global.DTInt16:  {2, 15, true, false},
		global.DTUint16: {2, 15, false, false},
		global.DTInt32:  {4, 31, true, false},
		global.DTUint32: {4, 31, false, false},
		global.DTInt64:  {8, 63, true, false},
		global.DTUint64: {8, 63, false, false},
	}
}

type PointResponse struct {
	TotalCount int      `json:"totalCount"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
	TotalPages int      `json:"totalPages"`
	Points     []*Point `json:"points"`
}

type Point struct {
	ID             string  `json:"id" gorm:"primaryKey"`
	FunctionCode   string  `json:"functionCode"` //功能码
	Address        string  `json:"address"`      //点位地址
	DataType       string  `json:"dataType"`     //数据类型
	Tag            string  `json:"tag"`
	LuaExpression  string  `json:"luaExpression"`  //lua表达式
	Description    string  `json:"description"`    //描述
	AlarmFlag      string  `json:"alarmFlag"`      //告警标志
	AlarmLevel     string  `json:"alarmLevel"`     //告警等级
	Multiplier     float64 `json:"multiplier"`     //倍率
	Unit           string  `json:"unit"`           //单位
	Priority       byte    `json:"priority"`       //优先级
	Endianness     string  `json:"endianness"`     //大小端
	BitCalculation string  `json:"bitCalculation"` //bit位计算
	StartBit       int     `json:"startBit"`       //起始bit
	EndBit         int     `json:"endBit"`         //结束bit
	StorageMethod  string  `json:"storageMethod"`  //存储方式，变存，直存
	Offset         float64 `json:"offset"`         //偏移量
	Store          int     `json:"store"`          //入库间隔秒
	DeviceID       string  `json:"deviceId"`       //设备id
}

func (p *Point) BaseParse(resp ...byte) (interface{}, error) {
	if p.Endianness != global.BigEndian && p.Endianness != global.LittleEndian {
		return nil, errors.New("endianness must be either LittleEndian, BigEndian")
	}

	switch p.DataType {
	case global.DTInt8, global.DTInt16, global.DTInt32, global.DTInt64,
		global.DTUint8, global.DTUint16, global.DTUint32, global.DTUint64:
		return p.numericCodec(resp)

	case global.DTFloat32:
		return p.float32Codec(resp)
	case global.DTFloat64:
		return p.float64Codec(resp)
	case global.DTInt8Arr:
		return p.int8ArrayCodec(resp)
	case global.DTInt16Arr:
		return p.int16ArrayCodec(resp)
	case global.DTInt32Arr:
		return p.int32ArrayCodec(resp)
	case global.DTInt64Arr:
		return p.int64ArrayCodec(resp)
	case global.DTUint8Arr:
		return p.uint8ArrayCodec(resp)
	case global.DTUint16Arr:
		return p.uint16ArrayCodec(resp)
	case global.DTUint32Arr:
		return p.uint32ArrayCodec(resp)
	case global.DTUint64Arr:
		return p.uint64ArrayCodec(resp)
	case global.DTFloat32Arr:
		return p.float32ArrayCodec(resp)
	case global.DTFloat64Arr:
		return p.float64ArrayCodec(resp)
	case global.DTString:
		return p.stringCodec(resp)
	default:
		return nil, errors.New("error data type")
	}
}

func (p *Point) numericCodec(resp []byte) (interface{}, error) {
	// 定义数据类型的信息
	info, exists := typeInfoMap[p.DataType]
	if !exists {
		return nil, fmt.Errorf("unsupported data type for numeric codec: %v", p.DataType)
	}
	// 检查数据长度
	if len(resp) < info.byteLen {
		return nil, fmt.Errorf("to %s, data length is < %d", p.DataType, info.byteLen)
	}
	// 转换原始值
	rawValue, err := p.convertRawValue(resp[:info.byteLen], info)
	if err != nil {
		return nil, err
	}
	// 处理位运算
	processedValue, err := p.processBitCalculation(rawValue, info.maxBit)
	if err != nil {
		return nil, err
	}
	// 转换为float64进行后续处理
	data := p.convertToFloat64(processedValue, info.signed)
	// 偏移量
	if p.Offset != 0 {
		data = data - p.Offset
	}
	// 倍率
	if p.Multiplier != 1 && p.Multiplier != 0 {
		data = data * p.Multiplier
	}
	// lua表达式
	return sl.execNumber(p.LuaExpression, data)
}

func (p *Point) convertRawValue(resp []byte, info dataTypeInfo) (uint64, error) {
	switch info.byteLen {
	case 1:
		return uint64(resp[0]), nil
	case 2:
		if p.Endianness == global.BigEndian {
			return uint64(binary.BigEndian.Uint16(resp)), nil
		}
		return uint64(binary.LittleEndian.Uint16(resp)), nil

	case 4:
		if p.Endianness == global.BigEndian {
			return uint64(binary.BigEndian.Uint32(resp)), nil
		}
		return uint64(binary.LittleEndian.Uint32(resp)), nil

	case 8:
		if p.Endianness == global.BigEndian {
			return binary.BigEndian.Uint64(resp), nil
		}
		return binary.LittleEndian.Uint64(resp), nil

	default:
		return 0, fmt.Errorf("unsupported byte length: %d", info.byteLen)
	}
}

func (p *Point) processBitCalculation(value uint64, maxBit int) (uint64, error) {
	if p.BitCalculation == global.AllBit {
		return value, nil
	}
	// 检查位范围
	if p.StartBit > maxBit {
		return 0, fmt.Errorf("start bit position out of range (0-%d)", maxBit)
	}
	if p.BitCalculation == global.SingleBit {
		// 单bit位
		return (value >> p.StartBit) & 1, nil
	}
	if p.BitCalculation == global.MultipleBit {
		// 多bit位
		if p.EndBit > maxBit {
			return 0, fmt.Errorf("end bit position out of range (0-%d)", maxBit)
		}
		if p.StartBit > p.EndBit {
			return 0, fmt.Errorf("start bit (%d) cannot be greater than end bit (%d)", p.StartBit, p.EndBit)
		}
		size := p.EndBit - p.StartBit + 1
		mask := (uint64(1) << size) - 1
		return (value >> p.StartBit) & mask, nil
	}
	return value, nil
}

func (p *Point) convertToFloat64(value uint64, signed bool) float64 {
	if !signed {
		return float64(value)
	}
	// 对于有符号类型，需要进行符号扩展
	switch p.DataType {
	case global.DTInt8:
		return float64(int8(value))
	case global.DTInt16:
		return float64(int16(value))
	case global.DTInt32:
		return float64(int32(value))
	case global.DTInt64:
		return float64(int64(value))
	default:
		return float64(value)
	}
}

func (p *Point) float32Codec(resp []byte) (interface{}, error) {
	if len(resp) < 4 {
		return nil, fmt.Errorf("to float32, error data length: %d", len(resp))
	}
	bits := binary.BigEndian.Uint32(resp[0:4])
	return math.Float32frombits(bits), nil
}

func (p *Point) float64Codec(resp []byte) (interface{}, error) {
	if len(resp) < 8 {
		return nil, fmt.Errorf("to float64, error data length: %d", len(resp))
	}
	bits := binary.BigEndian.Uint64(resp[0:8])
	return math.Float64frombits(bits), nil
}

func (p *Point) int8ArrayCodec(resp []byte) (interface{}, error) {
	return resp, nil
}

func (p *Point) int16ArrayCodec(resp []byte) (interface{}, error) {
	if p.Endianness == global.BigEndian {
		return p.bytesToInt16Array(resp, binary.BigEndian)
	}
	return p.bytesToInt16Array(resp, binary.LittleEndian)
}

func (p *Point) bytesToInt16Array(data []byte, byteOrder binary.ByteOrder) ([]int16, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("byte array length must be multiple of 2, got %d", len(data))
	}

	result := make([]int16, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		// 将2个字节转换为int16
		result[i/2] = int16(byteOrder.Uint16(data[i : i+2]))
	}
	return result, nil
}

func (p *Point) int32ArrayCodec(resp []byte) (interface{}, error) {
	if p.Endianness == global.BigEndian {
		return p.bytesToInt32Array(resp, binary.BigEndian)
	}
	return p.bytesToInt32Array(resp, binary.LittleEndian)
}

func (p *Point) bytesToInt32Array(data []byte, byteOrder binary.ByteOrder) ([]int32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("byte array length must be multiple of 4, got %d", len(data))
	}

	result := make([]int32, len(data)/4)
	for i := 0; i < len(data); i += 4 {
		result[i/4] = int32(byteOrder.Uint32(data[i : i+4]))
	}
	return result, nil
}

func (p *Point) int64ArrayCodec(resp []byte) (interface{}, error) {
	if p.Endianness == global.BigEndian {
		return p.bytesToInt64Array(resp, binary.BigEndian)
	}
	return p.bytesToInt64Array(resp, binary.LittleEndian)
}

func (p *Point) bytesToInt64Array(data []byte, byteOrder binary.ByteOrder) ([]int64, error) {
	if len(data)%8 != 0 {
		return nil, fmt.Errorf("byte array length must be multiple of 8, got %d", len(data))
	}

	result := make([]int64, len(data)/8)
	for i := 0; i < len(data); i += 8 {
		result[i/8] = int64(byteOrder.Uint64(data[i : i+8]))
	}
	return result, nil
}

func (p *Point) uint8ArrayCodec(resp []byte) (interface{}, error) {
	return resp, nil
}

func (p *Point) uint16ArrayCodec(resp []byte) (interface{}, error) {
	if p.Endianness == global.BigEndian {
		return p.bytesToUInt16Array(resp, binary.BigEndian)
	}
	return p.bytesToUInt16Array(resp, binary.LittleEndian)
}

func (p *Point) bytesToUInt16Array(data []byte, byteOrder binary.ByteOrder) ([]uint16, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("byte array length must be multiple of 2, got %d", len(data))
	}

	result := make([]uint16, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		// 将2个字节转换为int16
		result[i/2] = byteOrder.Uint16(data[i : i+2])
	}
	return result, nil
}

func (p *Point) uint32ArrayCodec(resp []byte) (interface{}, error) {
	if p.Endianness == global.BigEndian {
		return p.bytesToUInt32Array(resp, binary.BigEndian)
	}
	return p.bytesToUInt32Array(resp, binary.LittleEndian)
}

func (p *Point) bytesToUInt32Array(data []byte, byteOrder binary.ByteOrder) ([]uint32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("byte array length must be multiple of 4, got %d", len(data))
	}

	result := make([]uint32, len(data)/4)
	for i := 0; i < len(data); i += 4 {
		result[i/4] = byteOrder.Uint32(data[i : i+4])
	}
	return result, nil
}

func (p *Point) uint64ArrayCodec(resp []byte) (interface{}, error) {
	if p.Endianness == global.BigEndian {
		return p.bytesToUInt64Array(resp, binary.BigEndian)
	}
	return p.bytesToUInt64Array(resp, binary.LittleEndian)
}

func (p *Point) bytesToUInt64Array(data []byte, byteOrder binary.ByteOrder) ([]uint64, error) {
	if len(data)%8 != 0 {
		return nil, fmt.Errorf("byte array length must be multiple of 8, got %d", len(data))
	}

	result := make([]uint64, len(data)/8)
	for i := 0; i < len(data); i += 8 {
		result[i/8] = byteOrder.Uint64(data[i : i+8])
	}
	return result, nil
}

func (p *Point) float32ArrayCodec(resp []byte) (interface{}, error) {
	if p.Endianness == global.BigEndian {
		return p.bytesToFloat32Array(resp, binary.BigEndian)
	}
	return p.bytesToFloat32Array(resp, binary.LittleEndian)
}

func (p *Point) bytesToFloat32Array(data []byte, byteOrder binary.ByteOrder) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("byte array length must be multiple of 4, got %d", len(data))
	}

	result := make([]float32, len(data)/4)
	for i := 0; i < len(data); i += 4 {
		// 将4个字节转换为uint32
		bits := byteOrder.Uint32(data[i : i+4])
		// 将uint32的位模式解释为float32
		result[i/4] = math.Float32frombits(bits)
	}
	return result, nil
}

func (p *Point) float64ArrayCodec(resp []byte) (interface{}, error) {
	if p.Endianness == global.BigEndian {
		return p.bytesToFloat64Array(resp, binary.BigEndian)
	}
	return p.bytesToFloat64Array(resp, binary.LittleEndian)
}

func (p *Point) bytesToFloat64Array(data []byte, byteOrder binary.ByteOrder) ([]float64, error) {
	if len(data)%8 != 0 {
		return nil, fmt.Errorf("byte array length must be multiple of 8, got %d", len(data))
	}

	result := make([]float64, len(data)/8)
	for i := 0; i < len(data); i += 8 {
		bits := byteOrder.Uint64(data[i : i+8])
		result[i/8] = math.Float64frombits(bits)
	}
	return result, nil
}

func (p *Point) stringCodec(resp []byte) (interface{}, error) {
	return hex.EncodeToString(resp), nil
}
