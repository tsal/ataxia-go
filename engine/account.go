package engine

/*
   Account structures and functions
*/

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/tsal/ataxia-go/connection"
	"github.com/tsal/ataxia-go/game"
	"io"
	"log"
	"strings"
)

// The Account struct defines each connected player at the engine level
type Account struct {
	ID         string
	Email      string
	Password   string
	Name       string
	Characters []string
	conn       *ataxiaConnection
	server     *Server
	character  *game.Character
	In         chan string
	Out        chan string
}

// The ataxiaConnection struct wraps all the lower-level networking details for each connected player
type ataxiaConnection struct {
	socket     io.ReadWriteCloser
	handler    connection.Handler
	remoteAddr string
	state      string
}

// NewAccount returns a pointer to a newly connected player
func NewAccount(server *Server, conn *ataxiaConnection) *Account {
	return &Account{
		ID:     uuid.New().String(),
		conn:   conn,
		server: server,
		In:     make(chan string, 1024),
		Out:    make(chan string, 1024),
		Name:   "Unknown",
	}
}

// Run is the main goroutine for players that handles login and sets up the input and output goroutines.
func (account *Account) Run() {
	buf := make([]byte, 1024)

	// Setup the account here.
	_, err := account.conn.handler.Write([]byte("Hello, welcome to Ataxia. What is your account name?\n"))
	if err != nil {
		log.Println(err)
		return
	}
	if _, err := account.conn.handler.Read(buf); err != nil {
		if err == io.EOF {
			log.Printf("read EOF, disconnecting anonymous player (%s)...", account.conn.remoteAddr)
			return
		} else {
			log.Println(err)
		}
		account.Close()
		return
	}
	_, err = account.conn.handler.Write([]byte(fmt.Sprintf("Hello %s.\n", string(buf))))
	if err != nil {
		log.Printf("lost player before greeting `%s` (%s)", account.Name, account.conn.remoteAddr)
		account.Close()
		return
	}
	account.Name = string(buf)
	log.Printf("account: player `%s` connected", account.Name)

	account.character = account.server.World.LoadCharacter(account.Name) // let them choose later
	account.character.Attach(account.In)
	account.server.AddPlayer(account)

	// Create an anonymous goroutine for reading
	go func() {
		for {
			if account.conn.socket == nil {
				return
			}

			data := make([]byte, 1024)
			n, err := account.Read(data)

			if err != nil {
				if err == io.EOF {
					log.Printf("account: read EOF, disconnecting player `%s`", account.Name)
				} else {
					log.Println(err)
				}
				account.Close()
				return
			}

			// TODO: Parse the command here
			if n > 0 {
				data = bytes.Trim(data, " \x00") // trim trailing space and nuls
				account.Parse(string(data))
				//				account.server.SendToAll(fmt.Sprintf("<%s> %s", account.Name, string(data)))
			}
		}
	}()

	// Create an anonymous goroutine for writing
	go func() {
		for line := range account.In {
			if account.conn.socket == nil {
				return
			}
			written := 0
			b := []byte(line)
			for written < len(line) {
				n, err := account.Write(b[written:])
				if err != nil {
					if err == io.EOF {
						log.Println("EOF on write, disconnecting account")
					} else {
						log.Println(err)
					}
					account.Close()
					return
				}
				written += n
			}
		}
	}()
}

// Close handles closing all the relevant structures for the player, such as their socket.
func (account *Account) Close() {
	if account.conn.socket != nil {
		account.conn.handler.Close()
		account.character.Detach()
		account.server.RemovePlayer(account)
		_ = account.conn.socket.Close()
		log.Println("connection closed:", account.Name)
	}
}

// Parse handles interpreting the player input
func (account *Account) Parse(input string) {
	log.Println("account-parse:", input)
	args := strings.Split(input, " ")
	ctx := context.WithValue(context.TODO(), "character", account.character.ID)
	err := account.character.World.CommandHandler.Handle(ctx, args...)
	if err != nil {
		log.Println("account-parse: error:", err)
		account.character.Write("Huh?\n")
	}
	account.character.Write("> ")
}

// Write to the player
func (account *Account) Write(buf []byte) (n int, err error) {
	if account.conn.socket == nil {
		return
	}

	return account.conn.handler.Write(buf)
}

// Read from the player
func (account *Account) Read(buf []byte) (n int, err error) {
	if account.conn.socket == nil {
		return
	}

	return account.conn.handler.Read(buf)
}
