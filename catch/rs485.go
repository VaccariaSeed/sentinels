package catch

import (
	"bufio"
	"context"
	"errors"
	"io"
	"sentinels/global"
	"sentinels/model"
	"sync"
	"syscall"
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
		c.ReadTimeout = time.Duration(R.WriteTimeout) * time.Millisecond
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

func (R *RS485Client) verifyExcept(err error) {
	if errors.Is(err, io.EOF) || errors.Is(err, syscall.EIO) {
		R.flushLinkedFlag(false)
		_ = R.Close()
		_ = R.Open()
	} else {
		R.doneFail()
	}
}

func (R *RS485Client) doneFail() {
	R.failNum++
	if R.failNum >= R.failSize {
		R.failNum = 0
		R.flushLinkedFlag(false)
		_ = R.Close()
		_ = R.Open()
	}
}

func (R *RS485Client) Close() error {
	R.lock.Lock()
	defer R.lock.Unlock()
	if R.conn != nil {
		err := R.conn.Close()
		if err == nil {
			R.conn = nil
			R.reader = nil
			R.flushLinkedFlag(false)
		}
		return err
	}
	return nil
}

func (R *RS485Client) Type() string {
	return global.RS485
}

func (R *RS485Client) Flush() error {
	return R.conn.Flush()
}

func (R *RS485Client) Write(data []byte) error {
	R.lock.Lock()
	defer R.lock.Unlock()
	_, err := R.conn.Write(data)
	return err
}

func (R *RS485Client) WriteByTimeout(timeout time.Duration, data []byte) error {
	R.lock.Lock()
	defer R.lock.Unlock()
	_ = R.Flush()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // 无论如何，确保取消上下文以释放资源
	resultChan := make(chan struct {
		n   int
		err error
	}, 1)
	go func() {
		n, err := R.conn.Write(data)
		resultChan <- struct {
			n   int
			err error
		}{n, err}
	}()
	select {
	case <-ctx.Done():
		// 超时发生
		return ctx.Err()
	case result := <-resultChan:
		// 写操作完成
		return result.err
	}
}

func (R *RS485Client) Read() ([]byte, error) {
	//R.lock.Lock()
	//defer R.lock.Unlock()
	//return R.pc.Decode(R.reader)
	return nil, nil
}

func (R *RS485Client) ReadByTimeout(_ time.Duration) ([]byte, error) {
	return R.Read()
}

func (R *RS485Client) SendAndWaitForReply(data []byte) ([]byte, error) {
	//R.lock.Lock()
	//defer R.lock.Unlock()
	//_ = R.Flush()
	//_, err := R.conn.Write(data)
	//if err != nil {
	//	return nil, err
	//}
	//return R.pc.Decode(R.reader)
	return nil, nil
}

func (R *RS485Client) SendAndWaitForReplyByTimeOut(data []byte, timeout time.Duration) ([]byte, error) {
	resultChan := make(chan struct {
		n   []byte
		err error
	}, 1)
	go func() {
		n, err := R.SendAndWaitForReply(data)
		resultChan <- struct {
			n   []byte
			err error
		}{n, err}
	}()
	select {
	case result := <-resultChan:
		return result.n, result.err
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}
}

func (R *RS485Client) Collect(key string, data []byte, point model.PointSnap) error {
	resp, err := R.SendAndWaitForReply(data)
	if err != nil {
		return err
	}
	return R.parse(resp, point)

}

// todo 没写完
func (R *RS485Client) parse(resp []byte, point model.PointSnap) error {
	if R.swap != nil {
		R.swap(R.Device, nil, 0)
	}
	return nil
}

func (R *RS485Client) Operate(opt *model.Operate) ([]byte, error) {
	var err error
	req, err := R.pc.Opt(opt.Cmd)
	if err != nil {
		return nil, err
	}
	if opt.ReplySize == 0 {
		opt.ReplySize = 1
	}
	var resp []byte
	for index := 0; index < opt.ReplySize; index++ {
		if opt.Timeout >= 0 {
			resp, err = R.SendAndWaitForReplyByTimeOut(req, time.Duration(opt.Timeout)*time.Millisecond)
		} else {
			resp, err = R.SendAndWaitForReply(req)
		}
		if err != nil {
			continue
		}
	}
	return resp, nil
}
