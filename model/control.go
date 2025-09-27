package model

import (
	"errors"
	"fmt"
	"math"
	"sentinels/global"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Operate struct {
	UniqueIdentifier string      `json:"uniqueIdentifier"` //命令唯一标志
	ReplySize        int         `json:"replySize"`        //重试次数
	SignType         string      `json:"signType"`         //标志类型，id or address
	Sign             string      `json:"sign"`             //根据SignType进行变化，如果SignType是id，此为id，如果SignType为address， 此为address
	SendTime         int64       `json:"sendTime"`         //命令发送的时间
	ValidityPeriod   int64       `json:"validityPeriod"`   //有效期，单位：毫秒， 当这个值加上命令发送时间大于当前毫米值，这命令执行
	Cmd              *OperateCmd `json:"cmd"`              //命令
}

func (o *Operate) Check() error {
	if o.ReplySize < 0 {
		o.ReplySize = 0
	}
	if o.SignType != global.LogoTypeId && o.SignType != global.LogoTypeAddress {
		return fmt.Errorf("not found this signType:%s", o.SignType)
	}
	if strings.TrimSpace(o.Sign) == "" {
		return errors.New("sign is empty")
	}
	if o.Cmd == nil {
		return errors.New("cmd is empty")
	}
	if o.SendTime > 0 && o.ValidityPeriod > 0 {
		if time.Now().UnixMilli() > o.SendTime+o.ValidityPeriod {
			return fmt.Errorf("command timed out, now:%d, deadline:%d", time.Now().UnixMilli(), o.SendTime+o.ValidityPeriod)
		}
	}
	return nil
}

type OperateCmd struct {
	Timeout  int64             `json:"timeout"`  //毫秒
	CmdType  string            `json:"cmdType"`  //命令类型,参照global.go中的【命令类型】
	FuncCode string            `json:"funcCode"` //功能码
	Value    map[string]string `json:"value"`
}

const (
	startAddrFlag = "startAddr"
	lengthFlag    = "length"
	valueFlag     = "value"
)

func (op *OperateCmd) IsModbusCmd() bool {
	var keyWord = []string{global.CopyRead, global.SetCmd}
	return slices.Contains(keyWord, op.CmdType)
}

func (op *OperateCmd) modbusItem(key string) (value uint16, err error) {
	if valueStr, ok := op.Value[key]; ok {
		valueUint, err := strconv.ParseUint(valueStr, 0, 16)
		if err != nil {
			return 0, err
		}
		if valueUint > math.MaxUint16 {
			return 0, fmt.Errorf("modbus copy read items error, %s value is too large", key)
		}
		value = uint16(valueUint)
		return value, nil
	}
	return 0, fmt.Errorf("modbus copy read items error, %s value is empty", key)
}

// ModbusCopyReadItems modbus抄读参数
// 起始位置，数量，错误
func (op *OperateCmd) ModbusCopyReadItems() (startAddr uint16, length uint16, err error) {
	if op.Value == nil || len(op.Value) < 2 {
		return 0, 0, errors.New("modbus copy read items error")
	}
	startAddr, err = op.ModbusStartAddress()
	if err != nil {
		return 0, 0, err
	}
	length, err = op.ModbusLength()
	return
}

func (op *OperateCmd) ModbusStartAddress() (startAddr uint16, err error) {
	return op.modbusItem(startAddrFlag)
}

func (op *OperateCmd) ModbusLength() (length uint16, err error) {
	return op.modbusItem(lengthFlag)
}

func (op *OperateCmd) ModbusSingleValue() (value uint16, err error) {
	return op.modbusItem(valueFlag)
}

func (op *OperateCmd) ModbusValueToBytes() (uint16, []byte, error) {
	if valueStr, ok := op.Value[valueFlag]; !ok {
		return 0, nil, errors.New("modbus cmd item:value error")
	} else {
		va := strings.Split(valueStr, ",")
		for _, v := range va {
			if v != "0" && v != "1" {
				return 0, nil, errors.New("modbus cmd item:value error")
			}
		}
		valueStr = strings.ReplaceAll(valueStr, ",", "")
		// 补全到8的倍数
		padding := (8 - len(valueStr)%8) % 8
		paddedStr := valueStr
		for i := 0; i < padding; i++ {
			paddedStr = "0" + paddedStr
		}

		result := make([]byte, len(paddedStr)/8)
		for i := 0; i < len(paddedStr); i += 8 {
			var b byte
			for j := 0; j < 8; j++ {
				b <<= 1
				if paddedStr[i+j] == '1' {
					b |= 1
				}
			}
			result[i/8] = b
		}
		return uint16(len(valueStr)), result, nil
	}

}

func (op *OperateCmd) ModbusMultipleValue() ([]byte, error) {
	if op.Value == nil || op.Value[valueFlag] == "" {
		return nil, errors.New("modbus copy read items error")
	}
	length, err := op.ModbusLength()
	if err != nil {
		return nil, err
	}
	values := strings.Split(op.Value[valueFlag], ",")
	if uint16(len(values)) != length {
		return nil, errors.New("modbus copy read items error, length not equal")
	}
	var result []byte
	for _, v := range values {
		vUint16, pe := strconv.ParseUint(v, 0, 16)
		if pe != nil {
			return nil, pe
		}
		result = append(result, byte(vUint16>>8), byte(vUint16))
	}
	return result, nil
}
