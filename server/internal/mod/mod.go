// Package mod provides the interface and tools of modules.
package mod

import (
	"fmt"
	"io"

	"server/internal/net"
)

// Executor manages the execution of commands.
type Executor interface {

	// Exec executes a command.
	Exec(cmd string, args []string) error

	// Cmds returns all commands supported by the executor.
	Cmds() []string
}

// Responder manages the response of packets.
type Responder interface {

	// Respond handles a packet received from the client.
	Respond(client net.Client, pkg *net.Packet) error

	// Packets returns all packet types supported by the responder.
	Packets() []net.PktType
}

// ID is used to uniquely specify a module.
type ID int

// Mod is the interface of a module.
type Mod interface {
	io.Closer

	fmt.Stringer

	// Executor is used to execute commands.
	Executor

	// Responder is used to handle packets.
	Responder

	// ID returns the module's ID.
	ID() ID

	// Name returns the module's name.
	Name() string

	// SetClient switches the current client.
	SetClient(client net.Client)
}

// CmdHandler is the handler of a command.
type CmdHandler func(args []string) error

// CmdHandlerMap is a map storing command handlers.
type CmdHandlerMap map[string]CmdHandler

// NetHandler is the handler of a packet.
type NetHandler func(client net.Client, pkg *net.Packet) error

// NetHandlerMap is a map storing packet handlers.
type NetHandlerMap map[net.PktType]NetHandler
