package game

import (
	"fmt"

	golua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	//	"log"
	"strconv"
	//	"github.com/xenith-studios/ataxia/lua"
)

// PublishAccessors publishes Go functions into Lua
func (world *World) PublishAccessors(state *golua.LState) {
	state.SetGlobal("SendToAll", luar.New(state, world.SendToAll))
	state.SetGlobal("SendToOthers", luar.New(state, world.SendToOthers))
	state.SetGlobal("SendToChar", luar.New(state, world.SendToChar))
	state.SetGlobal("GetCharacterData", luar.New(state, world.GetCharacterData))
	state.SetGlobal("SetCharacterData", luar.New(state, world.SetCharacterData))
	state.SetGlobal("GetRoomData", luar.New(state, world.GetRoomData))
	state.SetGlobal("GetRoomExit", luar.New(state, world.GetRoomExit))
	state.SetGlobal("GetRoomExitData", luar.New(state, world.GetRoomExitData))
	state.SetGlobal("GetObjectData", luar.New(state, world.GetObjectData))
	state.SetGlobal("SetObjectData", luar.New(state, world.SetObjectData))
	state.SetGlobal("GetDictData", luar.New(state, world.GetDictData))
	state.SetGlobal("TestList", luar.New(state, TestList))
}

// TestList does something ##TODO
func TestList() map[string]string {
	m := make(map[string]string)
	m["1"] = "foo"
	m["2"] = "bar"
	m["3"] = "fubar"
	return m
}

// SendToAll sends to all connected characters
func (world *World) SendToAll(msg string) {
	msg = fmt.Sprintf("%s\n", msg)
	for _, ch := range world.Characters {
		ch.Write(msg)
		// if ch.Player != nil {
		// 	log.Println(msg)
		// 	ch.Player.In <- fmt.Sprintf("%s\r\n", msg)
		// }
	}
}

// SendToOthers sends to all other characters (aside from the specified character)
func (world *World) SendToOthers(charID string, msg string) {
	for id, ch := range world.Characters {
		if id == charID {
			continue
		}

		ch.Write(fmt.Sprintf("%s\n", msg))
		// if ch.Player != nil {
		// 	log.Println(msg)
		// 	ch.Player.In <- fmt.Sprintf("%s\r\n", msg)
		// }
	}
}

// SendToChar sends to the specified character
func (world *World) SendToChar(id string, msg string) {
	ch := world.Characters[id]
	if ch != nil {
		ch.Write(msg)
		// if ch.Player != nil {
		// 	ch.Player.In <- msg
		// }
	}
}

// GetObjectData ##TODO
func (world *World) GetObjectData(id string, field string) (ret string) {
	if world.Characters[id] != nil {
		return world.GetCharacterData(id, field)
	}
	if world.Rooms[id] != nil {
		return world.GetRoomData(id, field)
	}
	if world.RoomExits[id] != nil {
		return world.GetRoomExitData(id, field)
	}
	return ""
}

// SetObjectData ##TODO
func (world *World) SetObjectData(id string, field string, value string) {
	if world.Characters[id] != nil {
		world.SetCharacterData(id, field, value)
	}
	//	if world.Rooms[id] != nil { world.SetRoomData(id, field, value) }
	//	if world.RoomExits[id] != nil { world.SetRoomExitData(id, field, value) }
}

// GetDictData ##TODO
func (world *World) GetDictData(id string, field string, key string) string {
	if world.Rooms[id] == nil {
		return ""
	}

	room := world.Rooms[id]
	if field == "exits" {
		keyv, _ := strconv.Atoi(key)
		exit := room.exits[keyv]
		if exit != nil {
			return exit.ID
		}
	}

	return ""
}

// GetCharacterData ##TODO
func (world *World) GetCharacterData(id string, field string) string {
	ch := world.Characters[id]
	if ch == nil {
		return ""
	}

	if field == "name" {
		return ch.Name
	}
	if field == "room" {
		return ch.Room.ID
	}
	return ""
}

// SetCharacterData ##TODO
func (world *World) SetCharacterData(id string, field string, value string) {
	ch := world.Characters[id]
	if ch == nil {
		return
	}

	if field == "room" {
		ch.Room = world.Rooms[value]
		return
	}
}

// GetRoomData ##TODO
func (world *World) GetRoomData(id string, field string) string {
	ch := world.Rooms[id]
	if ch == nil {
		return ""
	}

	if field == "name" {
		return ch.Name
	}
	if field == "description" {
		return ch.Description
	}
	return ""
}

// GetRoomExit ##TODO
func (world *World) GetRoomExit(roomID string, dir int) string {
	room := world.Rooms[roomID]
	if room == nil {
		return ""
	}

	if room.exits[dir] != nil {
		return room.exits[dir].ID
	}

	return ""
}

// GetRoomExitData ##TODO
func (world *World) GetRoomExitData(exitID string, field string) string {
	exit := world.RoomExits[exitID]
	if exit == nil {
		return ""
	}

	if field == "destination" {
		if exit.Destination == nil {
			return ""
		}

		return exit.Destination.ID
	}
	return ""
}
