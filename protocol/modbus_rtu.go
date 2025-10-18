package protocol

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sentinels/command"
	"sentinels/global"
	"sentinels/snap"
	"strconv"
	"strings"
)

const (
	mrReadCoils              byte = 0x01 //读线圈,位,取得一组逻辑线圈的当前状态(ON/OFF)
	mrReadDiscreteInputs     byte = 0x02 //读离散输入寄存器,位,取得一组开关输入的当前状态(ON/OFF)
	mrReadHoldingRegisters   byte = 0x03 //读保持寄存器,整型、浮点型、字符型,在一个或多个保持寄存器中取得当前的二进制值
	mrReadInputRegisters     byte = 0x04 //读输入寄存器,整型、浮点型,在一个或多个输入寄存器中取得当前的二进制值
	mrWriteSingleCoil        byte = 0x05 //写单个线圈寄存器,位,强置一个逻辑线圈的通断状态
	mrWriteSingleRegister    byte = 0x06 //写单个保持寄存器,整型、浮点型、字符型,把具体二进制值装入一个保持寄存器
	mrWriteMultipleCoils     byte = 0x0F //写多个线圈寄存器,位,强置一串连续逻辑线圈的通断
	mrWriteMultipleRegisters byte = 0x10 //写多个保持寄存器,整型、浮点型、字符型,把具体的二进制值装入一串连续的保持寄存器
)

var _ ProtoConvener = (*ModbusRTU)(nil)

func init() {
	ProtoBuilder[global.ModbusRTU] = func(id string) (ProtoConvener, error) {
		slaveId, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
		if slaveId < 1 || slaveId > 247 {
			return nil, errors.New("invalid id " + id)
		}
		return &ModbusRTU{slaveId: byte(slaveId), proTool: &proTool{}}, nil
	}
}

type ModbusRTU struct {
	*proTool
	slaveId      byte
	funcCode     byte
	startAddress []byte
	wData        []uint16
	length       byte
}

func (m *ModbusRTU) Encode() ([]byte, error) {
	frame := []byte{m.slaveId, m.funcCode}
	frame = append(frame, m.startAddress...)
	if m.funcCode == mrReadCoils || m.funcCode == mrReadDiscreteInputs || m.funcCode == mrReadHoldingRegisters || m.funcCode == mrReadInputRegisters {
		frame = append(frame, 0, m.length)
	} else if m.funcCode == mrWriteSingleCoil || m.funcCode == mrWriteSingleRegister {
		frame = append(frame, byte(m.wData[0]>>8), byte(m.wData[0]))
	} else if m.funcCode == mrWriteMultipleRegisters {
		frame = append(frame, 0, m.length, byte(m.length*2))
		for _, d := range m.wData {
			frame = append(frame, byte(d>>8), byte(d))
		}
	} else if m.funcCode == mrWriteMultipleCoils {
		byteCount := (len(m.wData) + 7) / 8
		result := make([]byte, byteCount)
		for i, bit := range m.wData {
			if bit == 1 {
				// 计算字节索引和位索引
				byteIndex := i / 8
				bitIndex := uint(i % 8)
				// 设置对应的位
				result[byteIndex] |= 1 << bitIndex
			}
		}
		frame = append(frame, byte(len(result)))
		frame = append(frame, result...)
	}
	frame = append(frame, m.cs(frame)...)
	return frame, nil
}

func (m *ModbusRTU) Decode(reader *bufio.Reader) (string, []byte, error) {
	//读取
	var slaveId byte
	err := binary.Read(reader, binary.BigEndian, &slaveId)
	if err != nil {
		return "", nil, err
	}
	if slaveId != m.slaveId {
		return "", nil, fmt.Errorf("slave id error, readed:%d, must be:%d", slaveId, m.slaveId)
	}
	//读取功能码
	err = binary.Read(reader, binary.BigEndian, &m.funcCode)
	if err != nil {
		return "", nil, err
	}
	result := []byte{m.slaveId, m.funcCode}
	var data []byte
	if m.funcCode == mrReadCoils || m.funcCode == mrReadDiscreteInputs {
		//读线圈
		//返回字节数
		err = binary.Read(reader, binary.BigEndian, &m.length)
		if err != nil {
			return "", nil, err
		}
		result = append(result, m.length)
		//读取字节
		data = make([]byte, m.length)
		err = binary.Read(reader, binary.BigEndian, &data)
		if err != nil {
			return "", nil, err
		}
		result = append(result, data...)
		var rcData []byte
		for _, d := range data {
			rcData = append(rcData, m.byteToBitsSlice(d)...)
		}
		data = rcData
	} else if m.funcCode == mrReadHoldingRegisters || m.funcCode == mrReadInputRegisters {
		//返回字节数
		err = binary.Read(reader, binary.BigEndian, &m.length)
		if err != nil {
			return "", nil, err
		}
		result = append(result, m.length)
		data = make([]byte, m.length)
		err = binary.Read(reader, binary.BigEndian, &data)
		if err != nil {
			return "", nil, err
		}
		result = append(result, data...)
	} else if m.funcCode == mrWriteSingleCoil {
		m.startAddress = make([]byte, 2)
		err = binary.Read(reader, binary.BigEndian, &m.startAddress)
		if err != nil {
			return "", nil, err
		}
		result = append(result, m.startAddress...)
		data = make([]byte, 2)
		err = binary.Read(reader, binary.BigEndian, &data)
		if err != nil {
			return "", nil, err
		}
		m.wData = []uint16{binary.BigEndian.Uint16(data[:])}
	} else {
		return "", nil, fmt.Errorf("invalid modbus function code:%d", m.funcCode)
	}
	//计算cs
	cs := make([]byte, 2)
	err = binary.Read(reader, binary.BigEndian, &cs)
	if err != nil {
		return "", nil, err
	}
	checkCs := m.cs(result)
	if checkCs[0] != cs[0] || checkCs[1] != cs[1] {
		return "", nil, errors.New("cs error")
	}
	return hex.EncodeToString(append(result, cs...)), data, nil
}

func (m *ModbusRTU) Opt(cmd *command.OperateCmd) (string, []byte, error) {
	if strings.TrimSpace(cmd.FuncCode) == "" {
		return "", nil, NotFoundThisFuncCode
	}
	fc, err := strconv.ParseUint(cmd.FuncCode, 0, 8)
	if err != nil {
		return "", nil, err
	}
	fcByte := byte(fc)
	m.funcCode = fcByte
	if cmd.CmdType == global.CopyRead {
		if fcByte != mrReadCoils && fcByte != mrReadDiscreteInputs && fcByte != mrReadHoldingRegisters && fcByte != mrReadInputRegisters {
			return "", nil, errors.New("modbus rtu func code error")
		}
		startAddr, addrLength, cre := cmd.ModbusCopyReadItems()
		if cre != nil {
			return "", nil, cre
		}
		m.startAddress = []byte{byte(startAddr >> 8), byte(startAddr)}
		m.length = byte(addrLength)
		frame, encodeErr := m.Encode()
		return m.Key(), frame, encodeErr
	} else if fcByte == mrWriteSingleCoil {
		address, mse := cmd.ModbusStartAddress()
		if mse != nil {
			return "", nil, mse
		}
		m.startAddress = []byte{byte(address >> 8), byte(address)}
		value, msv := cmd.ModbusSingleValue()
		if msv != nil {
			return "", nil, msv
		}
		m.wData = []uint16{value}
		frame, encodeErr := m.Encode()
		return m.Key(), frame, encodeErr
	} else {
		return "", nil, errors.New("modbus rtu func code error")
	}
}

func (m *ModbusRTU) BuildBySnap(snap snap.PointSnap) (string, []byte, error) {
	m.funcCode = snap.FunctionCode()[0]
	m.startAddress = snap.Address()
	m.length = snap.Length()
	frame, err := m.Encode()
	return m.Key(), frame, err
}

func (m *ModbusRTU) CheckResp(frame, resp []byte) error {
	//TODO implement me
	panic("implement me")
}

func (m *ModbusRTU) Key() string {
	return fmt.Sprintf("modbusRTU_%d_%s_%d", m.funcCode, hex.EncodeToString(m.startAddress), m.length)
}

func (m *ModbusRTU) Copy() ProtoConvener {
	return &ModbusRTU{slaveId: m.slaveId}
}

func (m *ModbusRTU) cs(data []byte) []byte {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 0x0001) != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc = crc >> 1
			}
		}
	}
	return []byte{byte(crc & 0xFF), byte(crc >> 8)}
}
