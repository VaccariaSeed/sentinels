package catch

import (
	"errors"
	"sentinels/command"
	"sentinels/model"
	"sentinels/protocol"
	"sentinels/snap"
	"sync"
	"time"

	"go.uber.org/zap"
)

var DisConnectedError = errors.New("disconnected")

type SuccessLinked func(dev *model.Device)

type FailLinked func(dev *model.Device, err error)

type SwapCallback func(dev *model.Device, data map[string]interface{}, ts int64)

type CollectPointsFailCallBack func(dev *model.Device, point snap.PointSnap, err error)

type Connector interface {
	Open() error
	Close() error
	Type() string //对应的连接类型
	Flush() error
	Write(data []byte) error
	WriteByTimeout(timeout time.Duration, data []byte) error //带超时的发送
	Read() ([]byte, error)
	ReadByTimeout(timeout time.Duration) ([]byte, error)         //带超时的读取
	SendAndWaitForReply(key string, data []byte) ([]byte, error) //发送并等待回复
	SendAndWaitForReplyByTimeOut(key string, data []byte, timeout time.Duration) ([]byte, error)
	AddSuccessLinkedCallBack(call SuccessLinked)                 //添加连接成功的回调
	AddFailLinkedCallBack(call FailLinked)                       //添加断开连接的回调
	AddSwapCallback(callback SwapCallback)                       //采集到数据后发送到这里
	AddProtocolCodec(pc protocol.ProtoConvener)                  //添加编解码器
	AddCollectPointFailCallback(cps CollectPointsFailCallBack)   //采集失败回调
	Collect(key string, data []byte, point snap.PointSnap) error //采集
	parse(resp []byte, point snap.PointSnap) error               //解析
	ObtainDevice() *model.Device
	flushLinkedFlag(flag bool)
	IsLinked() bool

	Operate(opt *command.OperateCmd) ([]byte, error) //控制
	AddLogger(logger *zap.SugaredLogger)
}

type CreateConnectorFun func(device *model.Device) Connector

var ConnectorBuilder = make(map[string]CreateConnectorFun)

type ConnSyllable struct {
	*model.Device
	sc         SuccessLinked
	fc         FailLinked
	swap       SwapCallback
	cps        CollectPointsFailCallBack
	isLinked   bool
	linkedLock sync.Mutex
	pc         protocol.ProtoConvener
	logger     *zap.SugaredLogger
}

func (c *ConnSyllable) AddSuccessLinkedCallBack(call SuccessLinked) {
	c.sc = call
}

func (c *ConnSyllable) AddFailLinkedCallBack(call FailLinked) {
	c.fc = call
}

func (c *ConnSyllable) AddSwapCallback(callback SwapCallback) {
	c.swap = callback
}

func (c *ConnSyllable) AddCollectPointFailCallback(cps CollectPointsFailCallBack) {
	c.cps = cps
}

func (c *ConnSyllable) flushLinkedFlag(flag bool) {
	c.linkedLock.Lock()
	defer c.linkedLock.Unlock()
	if flag {
		if c.sc != nil {
			c.sc(c.Device)
		}
	} else {
		if c.fc != nil {
			c.fc(c.Device, DisConnectedError)
		}
	}
	c.isLinked = flag
}

func (c *ConnSyllable) IsLinked() bool {
	c.linkedLock.Lock()
	defer c.linkedLock.Unlock()
	return c.isLinked
}

func (c *ConnSyllable) AddProtocolCodec(pc protocol.ProtoConvener) {
	c.pc = pc
}

func (c *ConnSyllable) ObtainDevice() *model.Device {
	return c.Device
}

func (c *ConnSyllable) AddLogger(logger *zap.SugaredLogger) {
	c.logger = logger
}
