package catch

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"sentinels/global"
	"sentinels/model"
	"strings"
	"sync"
	"syscall"
	"time"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = l.Close()
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}

var _ Connector = (*TcpClient)(nil)

func init() {
	ConnectorBuilder[global.TcpClient] = func(device *model.Device) Connector {
		return &TcpClient{
			ConnSyllable: &ConnSyllable{Device: device},
			clientType:   global.TcpClient,
			bq:           model.NewBufQueue(50),
		}
	}
	ConnectorBuilder[global.TcpClientReuse] = func(device *model.Device) Connector {
		return &TcpClient{
			ConnSyllable: &ConnSyllable{Device: device},
			reuse:        true,
			clientType:   global.TcpClientReuse,
			bq:           model.NewBufQueue(50),
		}
	}
}

type TcpClient struct {
	*ConnSyllable
	reuse      bool //会进行端口复用
	localPort  int  //端口
	clientType string
	bq         *model.BufQueue
	conn       net.Conn
	reader     *bufio.Reader
	ctx        context.Context
	cancel     context.CancelFunc
	transfer   sync.Map
}

func (t *TcpClient) Open() error {
	var err error
	if t.reuse {
		if t.localPort == 0 {
			t.localPort, err = getFreePort()
			if err != nil {
				t.fc(t.Device, err)
				return err
			}
		}
		dialer := &net.Dialer{
			Control: func(network, address string, c syscall.RawConn) error {
				return c.Control(func(fd uintptr) {
					// 设置 SO_REUSEADDR
					err = syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
					if err != nil {
						return
					}
				})
			},
			LocalAddr: &net.TCPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: t.localPort,
			},
		}
		t.conn, err = dialer.Dial("tcp", t.Device.Address)
	} else {
		//不进行端口复用
		t.conn, err = net.Dial("tcp", t.Device.Address)
	}
	if err != nil {
		t.fc(t.Device, err)
		return err
	} else {
		t.flushLinkedFlag(true)
		t.reader = bufio.NewReader(t.conn)
		go func() {
			_, err = t.Read()
			if err != nil && err == io.EOF {
				_ = t.Close()
				return
			}
		}()
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())
	return err
}

func (t *TcpClient) Close() error {
	err := t.conn.Close()
	if err == nil {
		t.cancel()
		t.flushLinkedFlag(false)
	}
	return err
}

func (t *TcpClient) Type() string {
	return t.clientType
}

func (t *TcpClient) Flush() error {
	for {
		_, err := io.ReadAll(t.conn)
		if err != nil {
			return nil
		}
	}
}

func (t *TcpClient) Write(data []byte) error {
	if t.WriteTimeout > 0 {
		_ = t.conn.SetWriteDeadline(time.Now().Add(time.Duration(t.WriteTimeout) * time.Second))
	}
	_, err := t.conn.Write(data)
	return err
}

func (t *TcpClient) WriteByTimeout(timeout time.Duration, data []byte) error {
	if timeout > 0 {
		_ = t.conn.SetWriteDeadline(time.Now().Add(timeout))
	}
	_, err := t.conn.Write(data)
	return err
}

func (t *TcpClient) Read() ([]byte, error) {
	for {
		select {
		case <-t.ctx.Done():
			return nil, t.ctx.Err()
		default:
			_ = t.conn.SetReadDeadline(time.Now().Add(time.Duration(t.ReadTimeout) * time.Second))
			frame, resp, size, err := t.pc.Decode(t.reader)
			if err != nil {
				if t.isDisConnected(err) {
					return nil, io.EOF
				}
				continue
			}
			t.logger.Debugf("received -> %s", frame)
			key := t.pc.Key()
			sch, ok := t.transfer.Load(key)
			if ok {
				if s, flag := sch.(*model.SCH); flag {
					s.Set(resp)
				}
				continue
			}
			ps := t.bq.Get(key)
			if ps == nil {
				continue
			}
			err = t.parse(resp, size, ps)
			if err != nil {
				t.cps(t.Device, ps, err)
			}
		}
	}
}

func (t *TcpClient) ReadByTimeout(timeout time.Duration) ([]byte, error) {
	_ = t.conn.SetReadDeadline(time.Now().Add(timeout))
	_, resp, _, err := t.pc.Decode(t.reader)
	return resp, err
}

func (t *TcpClient) SendAndWaitForReply(key string, data []byte) ([]byte, error) {
	return t.SendAndWaitForReplyByTimeOut(key, data, global.DefaultTimeout)
}

func (t *TcpClient) SendAndWaitForReplyByTimeOut(key string, data []byte, timeout time.Duration) ([]byte, error) {
	t.logger.Debugf("send -> %s", hex.EncodeToString(data))
	sch := model.NewSCH(timeout)
	defer t.transfer.Delete(key)
	defer sch.Close()
	t.transfer.Store(key, sch)
	err := t.Write(data)
	if err != nil {
		return nil, err
	}
	err = sch.Wait()
	if err != nil {
		return nil, err
	}
	return sch.GetBytes()
}

func (t *TcpClient) Collect(key string, data []byte, point model.PointSnap) error {
	_ = t.conn.SetWriteDeadline(time.Now().Add(time.Duration(t.WriteTimeout) * time.Second))
	t.bq.Add(key, point)
	t.logger.Debugf("send -> %s", hex.EncodeToString(data))
	_, err := t.conn.Write(data)
	if err != nil {
		_ = t.bq.Get(key)
		return err
	}
	return nil
}

func (t *TcpClient) parse(resp []byte, size int, point model.PointSnap) error {
	if resp == nil || len(resp) == 0 {
		return errors.New("empty response")
	}
	result, err := point.Parse(resp, size)
	if err != nil {
		return err
	}
	t.swap(t.Device, result, time.Now().UnixMilli())
	return nil
}

func (t *TcpClient) isDisConnected(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	}
	// 检查系统调用错误：连接重置（Windows和Linux）
	if errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	// 检查系统调用错误：连接中止（Windows和Linux）
	if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNABORTED) {
		return true
	}
	// 检查网络操作错误
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		// 检查底层系统调用错误
		if opErr.Err != nil {
			errMsg := strings.ToLower(opErr.Err.Error())
			disconnectKeywords := []string{
				"connection reset",
				"broken pipe",
				"wsaeconnreset", // Windows specific
				"forcibly closed",
				"closed by peer",
			}
			for _, keyword := range disconnectKeywords {
				if strings.Contains(errMsg, keyword) {
					return true
				}
			}
		}
	}
	// 检查错误消息中的常见断开连接关键词
	errMsg := strings.ToLower(err.Error())
	disconnectKeywords := []string{
		"connection reset",
		"broken pipe",
		"wsaeconnreset",
		"forcibly closed",
		"closed by peer",
		"use of closed network connection",
	}
	for _, keyword := range disconnectKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	return false
}

func (t *TcpClient) Operate(ti []byte, opt *model.OperateCmd) ([]byte, error) {
	//生成报文
	pc := t.pc.Copy()
	key, frame, err := pc.Opt(ti, opt)
	if err != nil {
		return nil, err
	}
	var resp []byte
	if opt.Timeout > 0 {
		resp, err = t.SendAndWaitForReplyByTimeOut(key, frame, time.Duration(opt.Timeout)*time.Millisecond)
	} else {
		resp, err = t.SendAndWaitForReply(key, frame)
	}
	if err != nil {
		return nil, err
	}
	return resp, pc.CheckResp(frame, resp)
}
