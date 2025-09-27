package catch

import (
	"bufio"
	"sentinels/global"
	"sentinels/model"
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

func (R *RS485Client) Close() error {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) Type() string {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) Flush() error {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) Write(data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) WriteByTimeout(timeout time.Duration, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) Read() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) ReadByTimeout(timeout time.Duration) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) SendAndWaitForReply(key string, data []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) SendAndWaitForReplyByTimeOut(key string, data []byte, timeout time.Duration) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) Collect(key string, data []byte, point model.PointSnap) error {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) parse(resp []byte, size int, point model.PointSnap) error {
	//TODO implement me
	panic("implement me")
}

func (R *RS485Client) Operate(ti []byte, opt *model.OperateCmd) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
