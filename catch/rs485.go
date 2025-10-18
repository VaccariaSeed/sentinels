package catch

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"math"
	"sentinels/command"
	"sentinels/global"
	"sentinels/model"
	"sentinels/snap"
	"sync"
	"time"

	"github.com/tarm/serial"
)

var _ Connector = (*RS485Client)(nil)

var paritySnap = make(map[string]serial.Parity)

func init() {
	paritySnap["N"] = serial.ParityNone
	paritySnap["O"] = serial.ParityOdd
	paritySnap["E"] = serial.ParityEven
	paritySnap["M"] = serial.ParityMark
	paritySnap["S"] = serial.ParitySpace
	ConnectorBuilder[global.RS485] = func(device *model.Device) Connector {
		return &RS485Client{
			ConnSyllable: &ConnSyllable{Device: device},
			failSize:     6,
		}
	}
}

// RS485Client RS485的连接器
type RS485Client struct {
	*ConnSyllable
	failSize int
	failNum  int
	conn     *serial.Port
	lock     sync.Mutex
	reader   *bufio.Reader
}

func (R *RS485Client) Open() error {
	R.lock.Lock()
	defer R.lock.Unlock()
	parity := paritySnap[R.Parity]
	c := &serial.Config{Name: R.Address, Baud: R.BaudRate, Size: byte(R.DataBits), Parity: parity, StopBits: serial.StopBits(R.StopBits)}
	if R.ReadTimeout >= 0 {
		c.ReadTimeout = time.Duration(R.WriteTimeout) * time.Second
	} else {
		c.ReadTimeout = global.DefaultTimeout
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		R.fc(R.Device, err)
		return err
	}
	R.conn = s
	R.reader = bufio.NewReader(R.conn)
	R.flushLinkedFlag(true)
	return nil
}

func (R *RS485Client) Close() error {
	R.lock.Lock()
	defer R.lock.Unlock()
	err := R.conn.Close()
	R.conn = nil
	R.reader = nil
	R.flushLinkedFlag(false)
	return err
}

func (R *RS485Client) Type() string {
	return global.RS485
}

func (R *RS485Client) Flush() error {
	return R.conn.Flush()
}

func (R *RS485Client) Write(data []byte) error {
	err := R.Flush()
	if err != nil {
		return err
	}
	_, err = R.conn.Write(data)
	return err
}

func (R *RS485Client) WriteByTimeout(timeout time.Duration, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	done := make(chan error, 1)

	go func() {
		_, err := R.conn.Write(data)
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err() // 返回超时错误
	}
}

func (R *RS485Client) Read() ([]byte, error) {
	frame, result, err := R.pc.Decode(R.reader)
	if err == nil {
		R.logger.Debugf("received -> %s", frame)
	}
	return result, err
}

func (R *RS485Client) ReadByTimeout(_ time.Duration) ([]byte, error) {
	return nil, errors.New("such method ReadByTimeout not import")
}

func (R *RS485Client) SendAndWaitForReply(key string, data []byte) ([]byte, error) {
	return R.SendAndWaitForReplyByTimeOut(key, data, 0)
}

func (R *RS485Client) SendAndWaitForReplyByTimeOut(_ string, data []byte, _ time.Duration) ([]byte, error) {
	R.lock.Lock()
	defer R.lock.Unlock()
	R.logger.Debugf("send -> %s", hex.EncodeToString(data))
	err := R.Write(data)
	if err != nil {
		return nil, err
	}
	//暂停时间根据波特率而定
	time.Sleep(R.calculateTimeout(len(data)))
	return R.Read()
}

// 计算暂停时间
func (R *RS485Client) calculateTimeout(length int) time.Duration {
	totalBits := length * 10
	transmissionTime := float64(totalBits) / float64(R.BaudRate)
	return time.Duration(math.Ceil(transmissionTime*1000)) * time.Millisecond
}

func (R *RS485Client) Collect(key string, data []byte, point snap.PointSnap) error {
	resp, err := R.SendAndWaitForReplyByTimeOut(key, data, 0)
	if err != nil {
		return err
	}
	return R.parse(resp, point)
}

func (R *RS485Client) parse(resp []byte, point snap.PointSnap) error {
	if resp == nil || len(resp) == 0 {
		return errors.New("empty response")
	}
	result, err := point.Parse(resp)
	if err != nil {
		return err
	}
	R.swap(R.Device, result, time.Now().UnixMilli())
	return nil
}

func (R *RS485Client) Operate(opt *command.OperateCmd) ([]byte, error) {
	key, frame, err := R.pc.Opt(opt)
	if err != nil {
		return nil, err
	}
	result, err := R.SendAndWaitForReply(key, frame)
	if err != nil {
		return nil, err
	}
	return result[1:], nil
}
