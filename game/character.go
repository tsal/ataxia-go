package game

import (
	"github.com/google/uuid"
	"github.com/tsal/ataxia-go/lua"
	goLua "github.com/yuin/gopher-lua"
)

// Character defines a single character
type Character struct {
	ID     string
	Name   string
	World  *World
	Room   *Room
	output chan string
}

func (ch *Character) PublishAccessors(state *goLua.LState) {
	panic("implement me")
}

func (ch *Character) Handler() lua.HandlerFunc {
	return ch.World.CommandHandler.Handler
}

// NewCharacter returns a new charcater
func NewCharacter(world *World) *Character {
	ch := Character{
		World: world,
		ID:    uuid.New().String(),
	}

	return &ch
}

// Write to the character
func (ch *Character) Write(str string) {
	if ch.output != nil {
		ch.output <- str
	}
}

// Attach an output channel to the character
func (ch *Character) Attach(c chan string) {
	ch.output = c
}

// Detach the character's output channel
func (ch *Character) Detach() {
	ch.output = nil
}

var _ lua.Accessor = new(Character)
