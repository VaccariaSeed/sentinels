package protocol

import (
	"bufio"
	"sentinels/global"
	"sentinels/model"
	"strconv"
)

var _ ProtoConvener = (*ModbusRTU)(nil)

func init() {
	ProtoBuilder[global.ModbusRTU] = func(id string) ProtoConvener {
		slaveId, _ := strconv.Atoi(id)
		return &ModbusRTU{slaveId: byte(slaveId)}
	}
}

type ModbusRTU struct {
	slaveId byte
}

func (m ModbusRTU) Encode() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) Decode(reader *bufio.Reader) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) Opt(cmd model.OperateCmd) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) BuildBySnap(snap model.PointSnap) (string, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) Frame() string {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) Key() string {
	//TODO implement me
	panic("implement me")
}
