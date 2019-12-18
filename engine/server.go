package engine

/*
   Server structures and functions
*/

import (
	"github.com/tsal/ataxia-go/connection"
	"io"
	"log"
	"net"
	//	"bufio"
	"sync"
	"time"
	//	"container/list"
	"github.com/tsal/ataxia-go/game"
	"github.com/tsal/ataxia-go/lua"
	goLua "github.com/yuin/gopher-lua"
)

// PlayerList maintains a list of connected player accounts
type PlayerList struct {
	players map[string]*Account
	mu      sync.RWMutex
}

// Server struct defines main engine data structure
type Server struct {
	socket     *net.TCPListener
	luaState   *goLua.LState
	World      *game.World
	PlayerList *PlayerList
	In         chan string
	shutdown   chan bool
}

// NewPlayerList returns a new PlayerList struct
func NewPlayerList() *PlayerList {
	return &PlayerList{players: make(map[string]*Account)}
}

// Add a new account to the PlayerList
func (list *PlayerList) Add(name string, player *Account) {
	list.mu.Lock()
	defer list.mu.Unlock()
	list.players[name] = player
}

// Delete an account from the PlayerList
func (list *PlayerList) Delete(name string) {
	list.mu.Lock()
	defer list.mu.Unlock()
	list.players[name] = nil
}

// Get and return a player account by exact name
func (list *PlayerList) Get(name string) *Account {
	list.mu.RLock()
	defer list.mu.RUnlock()
	return list.players[name]
}

// NewServer creates a new server and returns a pointer to it
func NewServer(port int, shutdown chan bool) *Server {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(""), Port: port, Zone: ""})
	if err != nil {
		log.Fatalln("Failed to create server:", err)
		return nil
	}

	// Initiliate engine-local/goroutine-local LuaState
	L := lua.NewState()

	server := &Server{
		luaState:   L,
		World:      game.NewWorld(L),
		PlayerList: NewPlayerList(),
		socket:     listener,
		shutdown:   shutdown,
		In:         make(chan string, 1024),
	}

	lua.Publish(server, server.luaState)
	//server.PublishAccessors(server.luaState)
	server.World.PublishAccessors(server.luaState)

	// At this point, server and world go functions have been published
	// to Lua, we can load up some libraries for scripting action
	if err := L.DoFile("scripts/interface/context.lua"); err != nil {
		log.Fatal(err)
	}
	if err := L.DoFile("scripts/interface/accessors.lua"); err != nil {
		log.Fatal(err)
	}
	if err := L.DoFile("scripts/interface/character.lua"); err != nil {
		log.Fatal(err)
	}
	if err := L.DoFile("scripts/interface/room.lua"); err != nil {
		log.Fatal(err)
	}
	if err := L.DoFile("scripts/commands/character_action.lua"); err != nil {
		log.Fatal(err)
	}

	return server
}

// InitializeWorld initializes a single World
func (server *Server) InitializeWorld() {
	server.World.LoadAreas()
	server.World.Initialize()

}

// Shutdown the server and disconnect all remaining player accounts
func (server *Server) Shutdown() {
	if server.socket != nil {
		server.SendToPlayers("Server is shutting down!")
		for _, player := range server.PlayerList.players {
			if player != nil {
				player.Close()
			}
		}
		_ = server.socket.Close()
		server.socket = nil
		lua.Shutdown(server.luaState)
	}
}

// Listen is the goroutine that accepts new player connections and creates the account structure
func (server *Server) Listen() {
	for {
		if server.socket == nil {
			log.Println("Server socket closed")
			server.shutdown <- true
			return
		}
		// TODO: Implement a channel stack for connections to allow for further decoupling of ataxiaConnection handlers
		//select {
		//
		//}
		conn, err := server.socket.Accept()
		if err != nil {
			log.Println("Failed to accept new ataxiaConnection:", err)
			continue
		} else {
			c := new(ataxiaConnection)
			c.remoteAddr = conn.RemoteAddr().String()
			c.socket = conn
			c.handler = connection.NewTelnetHandler(conn)
			log.Println("server: new ataxiaConnection:", c.remoteAddr)
			player := NewAccount(server, c)
			go player.Run()
		}
	}
}

// Run is the main goroutine for the engine. It handles all game updates and events, also known as the main loop.
func (server *Server) Run() {
	// Main loop
	// Parse network messages (push user events)
	// Parse game updates
	// Game tick
	// Time update
	// Weather update
	// Entity updates (push events)
	// Parse pending events
	// Parse pending messages (network and player)
	// Sleep
	for {
		// Sleep for 1 ms
		time.Sleep(time.Millisecond)
	}
}

// AddPlayer adds a new player to the PlayerList
func (server *Server) AddPlayer(player *Account) {
	server.PlayerList.Add(player.Name, player)
}

// RemovePlayer removes a player from the PlayerList
func (server *Server) RemovePlayer(player *Account) {
	server.PlayerList.Delete(player.Name)
}

// Write is a convenience function to satisfy the io.Writer interface
func (server *Server) Write([]byte) (int, error) {
	return 0, io.EOF
}
