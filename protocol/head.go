package protocol

import (
	"bufio"
	"sentinels/command"
	"sentinels/snap"
)

type ProtoConvener interface {
	Encode() ([]byte, error)                             //编码
	Decode(reader *bufio.Reader) (string, []byte, error) //解码
	Opt(cmd *command.OperateCmd) (string, []byte, error)
	BuildBySnap(snap snap.PointSnap) (string, []byte, error)
	CheckResp(frame, resp []byte) error
	Key() string
	Copy() ProtoConvener
}

type ProtoCreateFunc func(id string) (ProtoConvener, error)

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

func (p *proTool) byteToBitsSliceAppendZero(b byte) []byte {
	bits := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bits[7-i] = byte((b >> i) & 1)
	}
	for i, j := 0, len(bits)-1; i < j; i, j = i+1, j-1 {
		bits[i], bits[j] = bits[j], bits[i]
	}
	result := make([]byte, 0, len(bits)*2)
	for _, bb := range bits {
		result = append(result, bb, 0x00)
	}
	return result
}
