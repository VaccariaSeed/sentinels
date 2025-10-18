package command

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sentinels/global"
	"strconv"
	"strings"
	"time"
)

const (
	startAddrFlag = "startAddr"
	lengthFlag    = "length"
	valueFlag     = "value"
)

// NewDefaultCarrier 创建一个“控制信息传输的载体”
func NewDefaultCarrier() *ControlCarrier {
	return &ControlCarrier{
		uniqueIdentifier: generateUUID(),
		ReplySize:        0,
		sendTime:         time.Now().UnixMilli(),
		ValidityPeriod:   time.Now().Add(10 * time.Second).UnixMilli(),
		Cmd:              &OperateCmd{Value: make(map[string]string)},
	}
}

// ControlCarrier 控制信息传输的载体
type ControlCarrier struct {
	uniqueIdentifier string      //唯一标识
	ReplySize        int         //重尝试次数
	signType         string      //根据什么寻找指定的调度器， address， id
	sign             string      //设备标识，与SignType相对应
	sendTime         int64       //发送时间
	ValidityPeriod   int64       //过期时间
	Cmd              *OperateCmd //参数
}

func (c *ControlCarrier) Check() error {
	if strings.TrimSpace(c.UniqueIdentifier()) == "" {
		return errors.New("unique identifier is empty")
	}
	if c.ReplySize < 0 {
		c.ReplySize = 0
	}
	if strings.TrimSpace(c.signType) == "" || strings.TrimSpace(c.sign) == "" {
		return errors.New("sign type or sign is empty")
	}
	if c.Cmd == nil {
		return errors.New("cmd is empty")
	}
	return c.Cmd.Check()
}

// ObtainSign 获取筛选标志
func (c *ControlCarrier) ObtainSign() (string, string) {
	return c.signType, c.sign
}

// 刷新设备筛选标识 编码
func (c *ControlCarrier) flushDevSignByTableFlag(flag string) *ControlCarrier {
	c.signType = global.LogoTypeTableFlag
	c.sign = flag
	return c
}

// 刷新设备筛选标识 id
func (c *ControlCarrier) flushDevSignByDevId(id string) *ControlCarrier {
	c.signType = global.LogoTypeId
	c.sign = id
	return c
}

// UniqueIdentifier 获取命令标识
func (c *ControlCarrier) UniqueIdentifier() string {
	return c.uniqueIdentifier
}

// FlushModbusCmdCopyRead 创建modbus的抄读命令
func (c *ControlCarrier) FlushModbusCmdCopyRead(funcCode byte, startAddress uint16, length uint16) *ControlCarrier {
	c.Cmd.CmdType = global.CopyRead
	c.Cmd.FuncCode = fmt.Sprintf("%d", funcCode)
	c.Cmd.Value[startAddrFlag] = fmt.Sprintf("%d", startAddress)
	c.Cmd.Value[lengthFlag] = fmt.Sprintf("%d", length)
	return c
}

// FlushModbusCmdSet 创建modbus的设置命令
func (c *ControlCarrier) FlushModbusCmdSet(funcCode byte, startAddress uint16, value ...uint16) *ControlCarrier {
	c.Cmd.CmdType = global.SetCmd
	c.Cmd.FuncCode = fmt.Sprintf("%d", funcCode)
	c.Cmd.Value[startAddrFlag] = fmt.Sprintf("%d", startAddress)
	c.Cmd.Value[lengthFlag] = fmt.Sprintf("%d", len(value))
	strSlice := make([]string, len(value))
	for i, v := range value {
		strSlice[i] = strconv.Itoa(int(v))
	}
	c.Cmd.Value[valueFlag] = strings.Join(strSlice, ",")
	return c
}

// FlushModbusCmdPassthrough 创建基于modbus的透传命令
func (c *ControlCarrier) FlushModbusCmdPassthrough(cmd []byte) *ControlCarrier {
	c.Cmd.CmdType = global.Passthrough
	c.Cmd.Value[valueFlag] = hex.EncodeToString(cmd)
	return c
}
