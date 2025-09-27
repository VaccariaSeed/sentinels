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

func (m ModbusRTU) Decode(reader *bufio.Reader) (string, []byte, int, error) {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) Opt(ti []byte, cmd *model.OperateCmd) (string, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) NextTi() []byte {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) BuildBySnap(snap model.PointSnap) (string, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) CheckResp(frame, resp []byte) error {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) Key() string {
	//TODO implement me
	panic("implement me")
}

func (m ModbusRTU) Copy() ProtoConvener {
	//TODO implement me
	panic("implement me")
}
