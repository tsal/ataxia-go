package lua

import (
	"context"
	"encoding/json"
	"fmt"
	goLua "github.com/yuin/gopher-lua"
	"io/ioutil"
	"log"
)

// Command defines a single command from a lua script
type Command struct {
	Script   string `json:"script"`
	FuncName string `json:"func_name"`
	Group    string `json:"group"`
}

type GoHandler interface {
	HandlerFunc() HandlerFunc
}

type HandlerFunc func(ctx context.Context, args ...string) (interface{}, error)

// CommandHandler defines a single command interpreter
type CommandHandler struct {
	commandList map[string]Command
	luaState    *goLua.LState
}

// NewCommandHandler returns a pointer to a new CommandHandler
func NewCommandHandler(luaState *goLua.LState) *CommandHandler {
	return &CommandHandler{
		luaState: luaState,
		// init stuff
	}
}

// LoadCommands loads all commands from the lua scripts as defined in the commands.json file
func (cmdHandler *CommandHandler) LoadCommands(commandFile string) {
	bytes, err := ioutil.ReadFile(commandFile)
	if err != nil {
		log.Fatal("Unable to read command list file.")
	}

	err = json.Unmarshal(bytes, &cmdHandler.commandList)
	if err != nil {
		log.Fatal("Unable to parse command list.")
	}

	for key := range cmdHandler.commandList {
		var err error
		// need to check and make sure a command with that name was loaded
		// should map these and only try and load lua scripts once, in case multiple commands
		// with same script file
		err = cmdHandler.luaState.DoFile(cmdHandler.commandList[key].Script)
		if err != nil {
			// Gracefully reject a script if it could not be loaded
			delete(cmdHandler.commandList, key)
			errMsg := fmt.Sprintf("skipping invalid command script: %s", err)
			err = nil // reset err
			log.Printf(errMsg)
		}
	}
	log.Printf("Loaded %d commands.", len(cmdHandler.commandList))
}

// Handler handles world commands until World has been refactored into an Accessor
func (cmdHandler *CommandHandler) Handler(ctx context.Context, args ...string) (interface{}, error) {
	// TODO: implement handler loop
	if c := ctx.Value("character"); c != nil {
		command, found := cmdHandler.commandList[args[0]]
		if !found {
			return 0, fmt.Errorf("command not found: %s", args[0])
		}
		fn := ExecuteCommand(cmdHandler.luaState, command)
		_, err := fn(ctx, args...)
		if err != nil {
			return -1, err
		}
		return 0, nil
	}

	return -1, fmt.Errorf("no handler for command: %s", args[0])
}
