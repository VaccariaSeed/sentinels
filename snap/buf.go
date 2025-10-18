package snap

import (
	"sync"
)

func NewBufQueue(size int) *BufQueue {
	return &BufQueue{
		size:    size,
		buf:     make(map[string]PointSnap),
		bufKeys: make([]string, size),
	}
}

// BufQueue 缓冲队列 先进先出
type BufQueue struct {
	size     int //大小
	swapSize int
	buf      map[string]PointSnap
	bufKeys  []string
	lock     sync.Mutex
}

func (t *BufQueue) Add(key string, point PointSnap) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.swapSize == t.size {
		//删除最左边
		firstKey := t.bufKeys[0]
		t.bufKeys = t.bufKeys[1:]
		delete(t.buf, firstKey)
	}
	t.swapSize++
	t.bufKeys = append(t.bufKeys, key)
	t.buf[key] = point
}

func (t *BufQueue) Get(key string) PointSnap {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.swapSize == 0 || len(t.buf) == 0 {
		return nil
	}
	size := 0
	i := 0
	for _, v := range t.bufKeys {
		if v != key {
			t.bufKeys[i] = v
			i++
		} else {
			size++
		}
	}
	t.bufKeys = t.bufKeys[:i]
	t.swapSize = t.swapSize - size
	ps := t.buf[key]
	delete(t.buf, key)
	return ps
}

// Size 获取当前长度
func (t *BufQueue) Size() int {
	return t.swapSize
}
