package task

import (
	"context"
	"errors"
	"io"
	"sentinels/catch"
	"sentinels/command"
	"sentinels/global"
	"sentinels/model"
	"sentinels/protocol"
	"sentinels/snap"
	"sync"
	"time"

	"go.uber.org/zap"
)

func NewGaTaskProcessor(device *model.Device) (*GaTaskProcessor, error) {
	//创建空调度器
	gtp := &GaTaskProcessor{}
	//获取调度器创建连接器的方法
	ctFunc, ok := catch.ConnectorBuilder[device.InterfaceType]
	if !ok {
		return nil, errors.New("connector not found for " + device.InterfaceType)
	}
	//创建连接器
	gtp.Connector = ctFunc(device)
	//获取编解码器
	ptFunc, ok := protocol.ProtoBuilder[device.ProtocolType]
	if !ok {
		return nil, errors.New("protocol not found for " + device.ProtocolType)
	}

	var err error
	gtp.Codec, err = ptFunc(device.DeviceAddress)
	if err != nil {
		return nil, err
	}
	gtp.Connector.AddProtocolCodec(gtp.Codec.Copy())
	//构建点位约束器
	gtp.pb, err = buildBinder(device)
	if err != nil {
		return nil, err
	}
	//创建日志组件
	gtp.logger = global.CreateLog(device.Identifier())
	gtp.Connector.AddLogger(gtp.logger)
	//添加各种回调
	gtp.AddSuccessLinkedCallBack(devConnected)
	gtp.AddFailLinkedCallBack(devDisConnected)
	gtp.AddSwapCallback(devSwap)
	gtp.AddCollectPointFailCallback(collectPointsFail)
	return gtp, nil
}

// GaTaskProcessor 采集控制调度器
type GaTaskProcessor struct {
	Connector catch.Connector        //连接器
	Codec     protocol.ProtoConvener //编解码器
	pb        *PointBinder           //点位集束器
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *zap.SugaredLogger
	lock      sync.Mutex
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
	var err error
	for {
		select {
		case <-g.ctx.Done():
			return nil
		default:
			err = g.collect(g.pb.Next())
			if err != nil && errors.Is(err, io.EOF) {
				return err
			}
		}
	}
}

func (g *GaTaskProcessor) collect(point snap.PointSnap) error {
	//剔除无用的点位
	if point == nil {
		return nil
	}
	//判断连接是否正常
	if !g.Connector.IsLinked() {
		g.logger.Error("collector not linked, device:", g.Connector.ObtainDevice().Identifier())
		return io.EOF
	}
	//根据点位组装报文
	key, frame, err := g.Codec.BuildBySnap(point)
	if err != nil {
		g.logger.Errorf("build by snap:\n%s \n err:%s", point.String(), err.Error())
	}
	//发送报文
	err = g.Connector.Collect(key, frame, point)
	if err != nil {
		g.logger.Errorf("read by snap:\n%s \n err:%s", point.String(), err.Error())
	}
	//暂停一段时间
	time.Sleep(time.Millisecond * 200 * 5 * 2) //暂停
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

func (g *GaTaskProcessor) operate(opt *command.OperateCmd) ([]byte, error) {
	if !g.Connector.IsLinked() {
		return nil, errors.New("connector not linked, device:" + g.Connector.ObtainDevice().Identifier())
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.Connector.Operate(opt)
}
