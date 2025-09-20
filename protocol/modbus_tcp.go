package protocol

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"sentinels/global"
	"sentinels/model"
	"slices"
	"strconv"
	"sync"
)

var ModbusTCPProtocolFlagError = errors.New("modbus tcp protocol flag error, must be [0x00, 0x00]")
var NotFoundThisFuncCode = errors.New("modbus tcp func code not found")

var _ ProtoConvener = (*ModbusTCP)(nil)

func init() {
	ProtoBuilder[global.ModbusTCP] = func(id string) ProtoConvener {
		slaveId, _ := strconv.Atoi(id)
		return &ModbusTCP{slaveId: byte(slaveId), protocolLogo: []byte{0x00, 0x00}, proTool: &proTool{}}
	}
}

var funcCodes = []byte{mtReadCoils, mtReadHoldingRegister}

const (
	mtReadCoils           byte = 0x01 //读线圈
	mtReadHoldingRegister      = 0x03
)

type ModbusTCP struct {
	*proTool
	slaveId      byte
	readSize     uint16 //读长度
	ti           uint16
	tiArr        []byte //事务标识符
	protocolLogo []byte //协议标志
	startAddress []byte
	funcCode     byte
	lock         sync.Mutex

	frame string
}

func (m *ModbusTCP) Frame() string {
	return m.frame
}

func (m *ModbusTCP) nextTi() []byte {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.ti++
	defer func() {
		if m.ti >= 0xff {
			m.ti = 0
		}
	}()
	return []byte{byte(m.ti >> 8), byte(m.ti)}
}

func (m *ModbusTCP) Encode() ([]byte, error) {
	//事务标识符,Modbus TCP协议标志
	result := append(m.tiArr, m.protocolLogo...)
	//后续数据
	data := append([]byte{m.slaveId, m.funcCode}, m.startAddress...)
	data = append(data, byte(m.readSize>>8), byte(m.readSize))
	//全拼
	result = append(result, byte(len(data)>>8), byte(len(data)))
	return append(result, data...), nil
}

func (m *ModbusTCP) Decode(reader *bufio.Reader) ([]byte, error) {
	peeked, err := reader.Peek(8)
	if err != nil {
		return nil, err
	}
	//获取Modbus TCP协议协议
	if peeked[2] != 0 && peeked[3] != 0 {
		_, _ = reader.ReadByte()
		return nil, ModbusTCPProtocolFlagError
	}
	//获取单元标识符
	ui := peeked[6]
	if ui != m.slaveId {
		_, _ = reader.ReadByte()
		return nil, errors.New("uint flag error")
	}
	//获取功能码
	fc := peeked[7]
	if !slices.Contains(funcCodes, fc) {
		_, _ = reader.ReadByte()
		return nil, errors.New("modbus tcp function code error")
	}
	//获取所有的长度
	frameSize := 8 - 2 + binary.BigEndian.Uint16([]byte{peeked[4], peeked[5]})
	var frame = make([]byte, frameSize)
	err = binary.Read(reader, binary.LittleEndian, &frame)
	if err != nil {
		return nil, err
	}
	m.frame = hex.EncodeToString(frame)
	m.tiArr = frame[0:2]
	m.funcCode = fc
	switch m.funcCode {
	case mtReadCoils:
		return m.decodeReadCoils(frame[9:])
	case mtReadHoldingRegister:
		return frame[9:], nil
	default:
		return nil, NotFoundThisFuncCode
	}
}

func (m *ModbusTCP) Opt(cmd model.OperateCmd) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *ModbusTCP) BuildBySnap(snap model.PointSnap) (string, []byte, error) {
	m.tiArr = m.nextTi()
	m.readSize = uint16(snap.Length())
	m.startAddress = snap.Address()
	m.funcCode = snap.FunctionCode()[0]
	frame, _ := m.Encode()
	return m.Key(), frame, nil
}

func (m *ModbusTCP) Key() string {
	return hex.EncodeToString(m.tiArr)
}

func (m *ModbusTCP) decodeReadCoils(bytes []byte) ([]byte, error) {
	var result []byte
	for _, b := range bytes {
		result = append(result, m.byteToBitsSlice(b)...)
	}
	return result, nil
}
