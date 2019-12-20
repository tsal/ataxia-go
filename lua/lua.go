package lua

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	goLua "github.com/yuin/gopher-lua"
)

// NewState returns a newly initalized LuaState
func NewState() *goLua.LState {
	log.Println("Initializing Lua State")
	st := goLua.NewState()
	return st
}

// Shutdown closes the LuaState
func Shutdown(st *goLua.LState) {
	if st != nil {
		st.Close()
	}
}

type Accessor interface {
	PublishAccessors(state *goLua.LState)
	Handler() HandlerFunc
}

func Publish(accessor Accessor, state *goLua.LState) {
	//state, lock := AcquireStateLock(state)
	//lock.Lock()
	//defer lock.Unlock()
	accessor.PublishAccessors(state)
}

// TODO: Further refactor
// ExecuteSingleArgPlayerCommand executes a single argument command, passes executing player id
func ExecuteCommand(st *goLua.LState, command Command) HandlerFunc {
	return func(ctx context.Context, args ...string) (i interface{}, err error) {
		if v := ctx.Value("character"); v != nil {
			c, ok := v.(string)
			if !ok {
				return -1, fmt.Errorf("invalid character ID passed to ExecuteSingleArgPlayerCommand")
			}
			actorID := c
			if len(args) < 1 {
				return 0, fmt.Errorf("no command passed")
			}
			funcName := command.FuncName
			// TODO: make use of luar
			stringArgs := ""
			if len(args) > 1 {
				stringArgs = strings.Join(args[1:], " ")
			}
			log.Printf("lua-execute: '%s' '%s' '%s'", actorID, funcName, stringArgs)
			err = st.CallByParam(goLua.P{
				Fn:      st.GetGlobal("execute_character_action"),
				NRet:    1,
				Protect: true,
			}, goLua.LString(actorID), goLua.LString(funcName), goLua.LString(stringArgs))
			if err != nil {
				log.Println("Lua script error in '", funcName, "' with args '", stringArgs, "':", err)
			}
			return 1, nil
		} else {
			// TODO: other handlers?
			return -1, fmt.Errorf("not implemented")
		}
	}
}

var locks map[*goLua.LState]*sync.Mutex

// AcquireStateLock implements a mutex as luar isn't thread-safe, so we need to protect mutator / accessor changes
func AcquireStateLock(st *goLua.LState) (state *goLua.LState, mutex *sync.Mutex) {
	if locks == nil {
		locks = make(map[*goLua.LState]*sync.Mutex)
		locks[st] = new(sync.Mutex)
	}
	if v, ok := locks[st]; ok {
		mutex = v
	} else {
		mutex = new(sync.Mutex)
		locks[st] = mutex
	}
	state = st
	mutex.Lock()
	return
}
