package model

import (
	"context"
	"errors"
	"sync"
	"time"
)

func NewSCH(timeout time.Duration) *SCH {
	s := &SCH{}
	s.waiter = timeout
	s.ch = make(chan interface{}, 1)
	return s
}

// SCH request后等待response同步通道
type SCH struct {
	value    interface{}
	ch       chan interface{}
	waiter   time.Duration
	lock     sync.Mutex
	isClosed bool
}

func (s *SCH) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.isClosed {
		return
	}
	close(s.ch)
	s.isClosed = true
}

func (s *SCH) Set(v interface{}) {
	select {
	case <-s.ch:
	default:
	}
	// 发送新值
	s.ch <- v
}

func (s *SCH) Wait() error {
	select {
	case value := <-s.ch:
		s.value = value
		return nil
	case <-time.After(s.waiter):
		return context.DeadlineExceeded
	}
}

func (s *SCH) GetBytes() ([]byte, error) {
	if s.value == nil {
		return nil, errors.New("value is nil")
	}
	if b, ok := s.value.([]byte); ok {
		return b, nil
	} else {
		return nil, errors.New("value is not []byte")
	}
}
