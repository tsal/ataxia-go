package lua

import (
	"log"
	"sync"

	golua "github.com/yuin/gopher-lua"
	//	luar "github.com/layeh/gopher-luar"
)

// NewState returns a newly initalized LuaState
func NewState() *golua.LState {
	log.Println("Initializing Lua State")
	st := golua.NewState()
	return st
}

// Shutdown closes the LuaState
func Shutdown(st *golua.LState) {
	if st != nil {
		st.Close()
	}
}

// Execute executes a simple command, one arg, no results
func Execute(st *golua.LState, funcName string, args string) {
	err := st.CallByParam(golua.P{
		Fn:      st.GetGlobal(funcName),
		NRet:    1,
		Protect: true,
	}, golua.LString(args))
	if err != nil {
		log.Println("Lua script error in '", funcName, "' with args '", args, "':", err)
	}
}

// ExecuteInterpret executes a two argument command, passes executing player id
func ExecuteInterpret(st *golua.LState, funcName string, actorID string, args string) {
	err := st.CallByParam(golua.P{
		Fn:      st.GetGlobal("execute_character_action"),
		NRet:    1,
		Protect: true,
	}, golua.LString(actorID), golua.LString(funcName), golua.LString(args))
	if err != nil {
		log.Println("Lua script error in '", funcName, "' with args '", args, "':", err)
	}
}

var locks map[*golua.LState]*sync.Mutex

// AcquireStateLock implements a mutex as luar isn't thread-safe, so we need to protect mutator / accessor changes
func AcquireStateLock(st *golua.LState) (state *golua.LState, mutex *sync.Mutex) {
	if locks == nil {
		locks = make(map[*golua.LState]*sync.Mutex)
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
