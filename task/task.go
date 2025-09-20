package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sentinels/catch"
	"sentinels/global"
	"sentinels/model"
	"sentinels/protocol"
	"sentinels/store"
	"time"

	"go.uber.org/zap"
)

func NewGaTaskProcessor(device *model.Device) (*GaTaskProcessor, error) {
	gtp := &GaTaskProcessor{}
	ctFunc, ok := catch.ConnectorBuilder[device.InterfaceType]
	if !ok {
		return nil, errors.New("connector not found for " + device.InterfaceType)
	}
	gtp.Connector = ctFunc(device)
	ptFunc, ok := protocol.ProtoBuilder[device.ProtocolType]
	if !ok {
		return nil, errors.New("protocol not found for " + device.ProtocolType)
	}
	gtp.Codec = ptFunc(device.DeviceAddress)
	gtp.Connector.AddProtocolCodec(ptFunc(device.DeviceAddress))
	//构建点位约束器
	var err error
	gtp.pb, err = buildBinder(device)
	if err != nil {
		return nil, err
	}
	gtp.logger = global.CreateLog(device.Identifier())
	gtp.Connector.AddLogger(gtp.logger)
	gtp.AddSuccessLinkedCallBack(devConnected)
	gtp.AddFailLinkedCallBack(devDisConnected)
	gtp.AddSwapCallback(devSwap)
	gtp.AddCollectPointFailCallback(collectPointsFail)
	return gtp, nil
}

var GTPSnapshot = make(map[string]*GaTaskProcessor)

func init() {
	devices := store.DbClient.SelectCutInDevice()
	if len(devices) == 0 {
		return
	}
	for _, device := range devices {
		gtp, err := NewGaTaskProcessor(devices[0])
		if err != nil {
			global.SystemLog.Error(fmt.Sprintf("id:%s name:%s type:%s err:%s", device.Id, device.Name, device.Code, err.Error()))
			continue
		}
		go func() {
			_ = gtp.Start()
		}()
		GTPSnapshot[device.Identifier()] = gtp
	}
}

// GaTaskProcessor 采集控制调度器
type GaTaskProcessor struct {
	Connector catch.Connector        //连接器
	Codec     protocol.ProtoConvener //编解码器
	pb        *PointBinder           //点位集束器
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *zap.SugaredLogger
}

func (g *GaTaskProcessor) Start() error {
	var err error
	//自动重连
	for {
		err = g.Connector.Open()
		if err != nil {
			time.Sleep(time.Millisecond * 1000 * 5)
			continue
		}
		break
	}
	g.ctx, g.cancel = context.WithCancel(context.Background())
	go func() {
		err = g.run()
		if err != nil {
			_ = g.Stop()
			go func() {
				_ = g.Start()
			}()
		}
	}()
	return nil
}

func (g *GaTaskProcessor) Stop() error {
	err := g.Connector.Close()
	g.cancel()
	return err
}

func (g *GaTaskProcessor) run() error {
	for {
		select {
		case <-g.ctx.Done():
			return nil
		default:
			err := g.collect(g.pb.Next())
			if err != nil && errors.Is(err, io.EOF) {
				return err
			}
		}
	}
}

func (g *GaTaskProcessor) collect(point model.PointSnap) error {
	if !g.Connector.IsLinked() {
		g.logger.Error("collector not linked, device:", g.Connector.ObtainDevice().Identifier())
		return io.EOF
	}
	key, frame, err := g.Codec.BuildBySnap(point)
	if err != nil {
		g.logger.Errorf("build by snap:\n%s \n err:%s", point.String(), err.Error())
	}
	err = g.Connector.Collect(key, frame, point)
	if err != nil {
		g.logger.Errorf("read by snap:\n%s \n err:%s", point.String(), err.Error())
	}
	time.Sleep(time.Millisecond * 200 * 5 * 2) //暂停两百毫秒
	return nil
}

func (g *GaTaskProcessor) ObtainDevice() *model.Device {
	return g.Connector.ObtainDevice()
}

func (g *GaTaskProcessor) AddSuccessLinkedCallBack(call catch.SuccessLinked) {
	g.Connector.AddSuccessLinkedCallBack(call)
}

func (g *GaTaskProcessor) AddFailLinkedCallBack(call catch.FailLinked) {
	g.Connector.AddFailLinkedCallBack(call)
}

func (g *GaTaskProcessor) AddSwapCallback(callback catch.SwapCallback) {
	g.Connector.AddSwapCallback(callback)
}

func (g *GaTaskProcessor) AddCollectPointFailCallback(cps catch.CollectPointsFailCallBack) {
	g.Connector.AddCollectPointFailCallback(cps)
}

func (g *GaTaskProcessor) Operate(opt *model.Operate) ([]byte, error) {
	return g.Connector.Operate(opt)
}
