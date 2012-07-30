/*
   Ataxia Mud Engine

   Copyright © 2009-2012 Xenith Studios
*/
package main

import (
	"./lua"
	"./settings"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

//	log "log4go.googlecode.com/hg"
)

// Variables for the command-line flags
var (
	portFlag       int
	configFlag     string
	hotbootFlag    bool
	descriptorFlag int
)

var shutdown chan bool

// Do all our basic initialization within the main package's init function.
func init() {
	fmt.Printf(`Ataxia Engine %s © 2009-2012, Xenith Studios (see AUTHORS)
Compiled on %s
Ataxia Engine comes with ABSOLUTELY NO WARRANTY; see COPYING for details.
This is free software, and you are welcome to redistribute it
under certain conditions; for details, see the file COPYING.

`, ATAXIA_VERSION, ATAXIA_COMPILED)

	shutdown = make(chan bool)
	// Setup the command-line flags
	flag.IntVar(&portFlag, "port", 0, "Main port")
	flag.StringVar(&configFlag, "config", "etc/config.lua", "Config file")
	flag.BoolVar(&hotbootFlag, "hotboot", false, "Recover from hotboot")
	flag.IntVar(&descriptorFlag, "descriptor", 0, "Hotboot descriptor")

	// Parse the command line
	flag.Parse()

	// Initialize Lua
	lua.Initialize()

	// Read configuration file
	ok := settings.LoadConfigFile(configFlag, portFlag)
	if !ok {
		log.Fatal("Error reading config file.")
	}

	// Initializations
	// Environment
	// Logging
	// Queues
	// Database

	if !hotbootFlag {
		// If previous shutdown was not clean and we are not recovering from a hotboot, clean up state and environment
	}
}

// When hotboot is called, this function will save game and world state, save each player state, and save the player list.
// Then it will do some cleanup (including closing the database) and call Exec to reload the running program.
func hotboot() {
	// Save game state
	// Save socket and player list
	// Disconnect from database
	arglist := append(os.Args, "-hotboot", "-descriptor=", fmt.Sprint(1234))
	syscall.Exec(os.Args[0], arglist, os.Environ())

	// If we get to this point, something went wrong. Die.
	log.Fatal("Failed to exec during hotboot.")
}

// When recovering from a hotboot, recover will restore the game and world state, restore the player list, and restore each player state.
// Once that is done, it will then reconnect each active descriptor to the associated player.
func recover() {
	log.Println("Recovering from hotboot.")
}

// 
func main() {
	// At this point, basic initialization has completed

	// Spin up a goroutine to handle signals
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
		for sig := range c {
			if usig, ok := sig.(os.Signal); ok {
				switch usig {
				case syscall.SIGQUIT:
					fallthrough
				case syscall.SIGTERM:
					fallthrough
				case syscall.SIGINT:
					// Catch the three interrupt signals and signal the game to shutdown.
					shutdown <- true
				case syscall.SIGHUP:
					// TODO: Reload settings and game state
				case syscall.SIGTSTP:
					// Pass on the SIGSTP signal through the syscall mechanism. This actually works.
					syscall.Kill(syscall.Getpid(), syscall.SIGSTOP)
				}
			}
		}
	}()

	// If configured, chroot into the designated directory
	if settings.Chroot != "" {
		err := syscall.Chroot(settings.Chroot)
		if err != nil {
			log.Fatalln("Failed to chroot:", err)
		}
		error := os.Chdir(settings.Chroot)
		if error != nil {
			log.Fatalln("Failed to chdir:", error)
		}
		log.Println("Chrooted to", settings.Chroot)
	}

	// Drop priviledges if configured

	// Daemonize if configured
	if settings.Daemonize {
		log.Println("Daemonizing")
		// Daemonize here
		// TODO: This probably won't be doable until Go supports forking into the background.
	}

	// Write out pid file
	pid := fmt.Sprint(os.Getpid())
	pidfile, err := os.Create(settings.Pidfile)
	if pidfile == nil {
		log.Fatalln("Error writing pid to file:", err)
	}
	pidfile.Write([]byte(pid))
	log.Println("Wrote PID to", settings.Pidfile)
	pidfile.Close()
	defer os.Remove(settings.Pidfile)

	// Initialize the network
	log.Println("Initializing network")
	server := NewServer(settings.MainPort, shutdown)

	// Initialize game state
	// Load database
	// Load commands
	// Load scripts
	// Load world
	// Load entities

	// Are we recovering from a hotboot?
	if hotbootFlag {
		recover()
	}

	// Initialization and setup is complete. Spin up a goroutine to handle incoming connections
	go server.Listen()

	// Run the game loop in its own goroutine
	go server.Run()

	// Wait for the shutdown signal
	<-shutdown

	// Cleanup
	log.Println("Cleaning up....")
	lua.Shutdown()
	server.Shutdown()
}
