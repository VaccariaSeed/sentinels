package protocol

import (
	"bufio"
	"sentinels/model"
)

type ProtoConvener interface {
	Encode() ([]byte, error)                     //编码
	Decode(reader *bufio.Reader) ([]byte, error) //解码
	Opt(cmd model.OperateCmd) ([]byte, error)
	BuildBySnap(snap model.PointSnap) (string, []byte, error)
	Frame() string
	Key() string
}

type ProtoCreateFunc func(id string) ProtoConvener

var ProtoBuilder = make(map[string]ProtoCreateFunc)

type proTool struct {
}

func (p *proTool) byteToBitsSlice(b byte) []byte {
	bits := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bits[7-i] = byte((b >> i) & 1)
	}
	for i, j := 0, len(bits)-1; i < j; i, j = i+1, j-1 {
		bits[i], bits[j] = bits[j], bits[i]
	}
	return bits
}
