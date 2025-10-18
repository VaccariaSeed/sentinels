package snap

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/yuin/gopher-lua"
)

var sl *sLua

func init() {
	sl = &sLua{needLock: false, lst: lua.NewState()}
}

type sLua struct {
	needLock bool
	lst      *lua.LState
	lock     sync.Mutex
}

func (s *sLua) close() {
	s.lst.Close()
}

func (s *sLua) execNumber(luaTemp string, value float64) (float64, error) {
	luaTemp = strings.TrimSpace(luaTemp)
	if luaTemp == "" {
		return value, nil
	}
	if s.needLock {
		s.lock.Lock()
		defer s.lock.Unlock()
	}
	// 设置全局变量
	s.lst.SetGlobal("value", lua.LNumber(value))
	// 确保有return语句
	if !strings.HasPrefix(luaTemp, "return ") {
		luaTemp = "return " + luaTemp
	}
	// 执行Lua代码
	err := s.lst.DoString(luaTemp)
	if err != nil {
		return 0, fmt.Errorf("lua execution error: %w", err)
	}
	// 检查栈顶是否有返回值
	top := s.lst.GetTop()
	if top == 0 {
		return 0, errors.New("no return value from lua expression")
	}
	// 获取返回值
	luaValue := s.lst.Get(-1) // 获取栈顶元素
	s.lst.Pop(1)              // 清理栈
	// 类型检查并转换
	if num, ok := luaValue.(lua.LNumber); ok {
		return float64(num), nil
	}
	return 0, fmt.Errorf("expected number return value, got %T", luaValue)
}
