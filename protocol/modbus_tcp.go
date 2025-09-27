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
	"strings"
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

var funcCodes = []byte{mtReadCoils, mtReadHoldingRegister, mtReadDiscreteInput, mtReadInputRegister, mtPreSetSingleCoil, mtWriteASingleHoldRegister, mtForceMultipleCoils, mtWriteMultipleHoldRegisters}

const (
	mtReadCoils                  byte = 0x01 //读线圈
	mtReadDiscreteInput          byte = 0x02 //读离散量输入
	mtReadHoldingRegister        byte = 0x03 //读取保存型寄存器
	mtReadInputRegister          byte = 0x04
	mtPreSetSingleCoil           byte = 0x05 //预置单线圈
	mtWriteASingleHoldRegister   byte = 0x06 //写单个保持寄存器
	mtForceMultipleCoils         byte = 0x0F //写多个线圈
	mtWriteMultipleHoldRegisters byte = 0x10 //写多个保持寄存器
)

type ModbusTCP struct {
	*proTool
	slaveId      byte
	readSize     uint16 //读长度
	data         []byte
	ti           uint16
	tiArr        []byte //事务标识符
	protocolLogo []byte //协议标志
	startAddress []byte
	funcCode     byte
	lock         sync.Mutex
}

func (m *ModbusTCP) Encode() ([]byte, error) {
	//事务标识符,Modbus TCP协议标志
	result := append(m.tiArr, m.protocolLogo...)
	//后续数据
	data := append([]byte{m.slaveId, m.funcCode}, m.startAddress...)
	data = append(data, byte(m.readSize>>8), byte(m.readSize))
	if m.data != nil {
		data = append(data, m.data...)
	}
	//全拼
	result = append(result, byte(len(data)>>8), byte(len(data)))
	return append(result, data...), nil
}

func (m *ModbusTCP) Decode(reader *bufio.Reader) (string, []byte, int, error) {
	peeked, err := reader.Peek(8)
	if err != nil {
		return "", nil, 0, err
	}
	//获取Modbus TCP协议协议
	if peeked[2] != 0 && peeked[3] != 0 {
		_, _ = reader.ReadByte()
		return "", nil, 0, ModbusTCPProtocolFlagError
	}
	//获取单元标识符
	ui := peeked[6]
	if ui != m.slaveId {
		_, _ = reader.ReadByte()
		return "", nil, 0, errors.New("uint flag error")
	}
	//获取功能码
	fc := peeked[7]
	if !slices.Contains(funcCodes, fc) {
		_, _ = reader.ReadByte()
		return "", nil, 0, errors.New("modbus tcp function code error")
	}
	//获取所有的长度
	frameSize := 8 - 2 + binary.BigEndian.Uint16([]byte{peeked[4], peeked[5]})
	var frame = make([]byte, frameSize)
	err = binary.Read(reader, binary.LittleEndian, &frame)
	if err != nil {
		return "", nil, 0, err
	}
	m.tiArr = frame[0:2]
	m.funcCode = fc
	switch m.funcCode {
	case mtReadCoils, mtReadDiscreteInput:
		resp, codecErr := m.decodeReadCoilsAndReadDiscreteInput(frame[9:])
		return hex.EncodeToString(frame), resp, 1, codecErr
	case mtReadHoldingRegister, mtReadInputRegister:
		return hex.EncodeToString(frame), frame[9:], 2, nil
	case mtPreSetSingleCoil, mtWriteASingleHoldRegister, mtForceMultipleCoils, mtWriteMultipleHoldRegisters:
		return hex.EncodeToString(frame), frame[8:], 1, nil
	default:
		return hex.EncodeToString(frame), nil, 0, NotFoundThisFuncCode
	}
}

func (m *ModbusTCP) Opt(ti []byte, cmd *model.OperateCmd) (string, []byte, error) {
	if !cmd.IsModbusCmd() {
		return "", nil, errors.New("modbus tcp command error")
	}
	if strings.TrimSpace(cmd.FuncCode) == "" {
		return "", nil, NotFoundThisFuncCode
	}
	fc, err := strconv.ParseUint(cmd.FuncCode, 0, 8)
	if err != nil {
		return "", nil, err
	}
	fcByte := byte(fc)
	if cmd.CmdType == global.CopyRead {
		if fcByte != mtReadCoils && fcByte != mtReadHoldingRegister && fcByte != mtReadDiscreteInput && fcByte != mtReadInputRegister {
			return "", nil, NotFoundThisFuncCode
		}
		//抄读
		startAddr, addrLength, cre := cmd.ModbusCopyReadItems()
		if cre != nil {
			return "", nil, cre
		}
		return m.buildFrame(addrLength, []byte{byte(startAddr >> 8), byte(startAddr)}, fcByte, ti, nil)
	} else {
		address, mse := cmd.ModbusStartAddress()
		if mse != nil {
			return "", nil, mse
		}
		//设置
		if fcByte == mtPreSetSingleCoil || fcByte == mtWriteASingleHoldRegister {
			value, msv := cmd.ModbusSingleValue()
			if msv != nil {
				return "", nil, msv
			}
			return m.buildFrame(value, []byte{byte(address >> 8), byte(address)}, fcByte, ti, nil)
		} else if fcByte == mtForceMultipleCoils {
			size, value, msv := cmd.ModbusValueToBytes()
			if msv != nil {
				return "", nil, msv
			}
			return m.buildFrame(size, []byte{byte(address >> 8), byte(address)}, fcByte, ti, append([]byte{byte(len(value))}, value...))
		} else if fcByte == mtWriteMultipleHoldRegisters {
			values, mve := cmd.ModbusMultipleValue()
			if mve != nil {
				return "", nil, mve
			}
			return m.buildFrame(uint16(len(values)/2), []byte{byte(address >> 8), byte(address)}, fcByte, ti, append([]byte{byte(len(values))}, values...))
		} else {
			return "", nil, NotFoundThisFuncCode
		}
	}

}

func (m *ModbusTCP) NextTi() []byte {
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

func (m *ModbusTCP) BuildBySnap(snap model.PointSnap) (string, []byte, error) {
	return m.buildFrame(uint16(snap.Length()), snap.Address(), snap.FunctionCode()[0], m.NextTi(), nil)
}

func (m *ModbusTCP) buildFrame(length uint16, address []byte, funcCode byte, ti []byte, data []byte) (string, []byte, error) {
	m.tiArr = ti
	m.readSize = length
	m.startAddress = address
	m.funcCode = funcCode
	m.data = data
	frame, _ := m.Encode()
	return m.Key(), frame, nil
}

func (m *ModbusTCP) CheckResp(frame, resp []byte) error {
	if m.funcCode == mtPreSetSingleCoil || m.funcCode == mtWriteASingleHoldRegister {
		if strings.HasSuffix(hex.EncodeToString(frame), hex.EncodeToString(resp)) {
			return nil
		}
	}
	if m.funcCode == mtForceMultipleCoils || m.funcCode == mtWriteMultipleHoldRegisters {
		if len(resp) == 4 && m.startAddress[0] == resp[0] && m.startAddress[1] == resp[1] && m.readSize == binary.BigEndian.Uint16(resp[2:4]) {
			return nil
		}
	}
	if m.funcCode == mtReadCoils || m.funcCode == mtReadDiscreteInput {
		return nil
	} else if m.funcCode == mtReadHoldingRegister || m.funcCode == mtReadInputRegister {
		return nil
	}
	return errors.New("modbus tcp resp error")
}

func (m *ModbusTCP) Key() string {
	return hex.EncodeToString(m.tiArr)
}

func (m *ModbusTCP) Copy() ProtoConvener {
	return &ModbusTCP{slaveId: m.slaveId, protocolLogo: []byte{0x00, 0x00}, proTool: &proTool{}}
}

func (m *ModbusTCP) decodeReadCoilsAndReadDiscreteInput(bytes []byte) ([]byte, error) {
	var result []byte
	for _, b := range bytes {
		result = append(result, m.byteToBitsSlice(b)...)
	}
	return result, nil
}
