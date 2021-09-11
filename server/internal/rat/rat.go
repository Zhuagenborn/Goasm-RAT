// Package rat provides the definition of remote administration tool.
package rat

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"server/internal/mod"
	ratnet "server/internal/net"
	"server/internal/utility/asynlog"
	"server/internal/utility/panic"
)

const (
	// Connect means the packet is related to network connections.
	Connect ratnet.PktType = 1
	// Disconnect means the packet is related to network disconnections.
	Disconnect ratnet.PktType = 2
)

// RAT is the interface of the remote administration tool.
type RAT interface {
	io.Closer

	// Register registers a module.
	Register(mod mod.Mod) error

	// Startup starts the remote administration tool.
	Startup(port int) error

	// Exec executes a command.
	Exec(cmd string, args []string) error
}

type rat struct {
	listener net.Listener

	sockets    *socketList
	currClient ratnet.Client

	abort     chan bool
	waitGroup sync.WaitGroup

	mod.Dispatcher
	cmdHandlers mod.CmdHandlerMap
	netHandlers mod.NetHandlerMap

	asynlog.Logger
}

// New creates a new remote administration tool.
func New(logger asynlog.Logger) RAT {
	rat := &rat{
		Logger:      logger,
		Dispatcher:  mod.NewDispatcher(),
		cmdHandlers: make(mod.CmdHandlerMap),
		netHandlers: make(mod.NetHandlerMap),
	}

	// Add command handlers and packet handlers.
	rat.cmdHandlers["exit"] = func([]string) error {
		return io.EOF
	}

	rat.cmdHandlers["sw"] = rat.switchClient
	rat.cmdHandlers["off"] = rat.terminateClient

	rat.netHandlers[Disconnect] = rat.onDisconnect
	rat.netHandlers[Connect] = func(ratnet.Client, *ratnet.Packet) error {
		return nil
	}

	if err := rat.Register(rat); err != nil {
		logger.Panic(err)
	}

	return rat
}

func (rat *rat) Startup(port int) error {
	var err error
	rat.listener, err = net.Listen("tcp", ":"+fmt.Sprintf("%v", port))
	if err != nil {
		return err
	}

	rat.sockets = newSocketList()
	rat.abort = make(chan bool)

	rat.waitGroup.Add(1)
	go rat.listen()
	return nil
}

func (rat *rat) Exec(cmd string, args []string) error {
	rat.LogStorage()
	if rat.currClient != nil && rat.sockets.Get(rat.currClient.ID()) == nil {
		rat.Log(fmt.Sprintf("The current client [%v] has become invalid.", rat.currClient.ID()))
		rat.SetClient(nil)
	}

	if cmd == "" {
		return nil
	}

	defer rat.LogStorage()

	mod := rat.ByCmd(cmd)
	if mod == nil {
		return fmt.Errorf("The command is invalid: %v", cmd)
	}

	// Give the command to the sub-module.
	if mod != rat {
		return mod.Exec(cmd, args)
	}

	// Handle commands supported by RAT itself.
	handler, ok := rat.cmdHandlers[cmd]
	panic.Assert(ok, "The module has registered an invalid command.")

	return handler(args)
}

func (rat *rat) Cmds() []string {
	cmds := make([]string, 0, len(rat.cmdHandlers))
	for c := range rat.cmdHandlers {
		cmds = append(cmds, c)
	}

	return cmds
}

// Respond only handles packets supported by RAT itself.
func (rat *rat) Respond(client ratnet.Client, pkg *ratnet.Packet) error {
	handler, ok := rat.netHandlers[pkg.Type]
	panic.Assert(ok, "The module has registered an invalid packet type.")

	return handler(client, pkg)
}

func (rat *rat) Packets() []ratnet.PktType {
	types := make([]ratnet.PktType, 0, len(rat.netHandlers))
	for t := range rat.netHandlers {
		types = append(types, t)
	}

	return types
}

func (*rat) ID() mod.ID {
	return 0
}

func (*rat) Name() string {
	return "RAT"
}

func (rat *rat) String() string {
	return fmt.Sprintf("%d-%s", rat.ID(), rat.Name())
}

func (rat *rat) SetClient(client ratnet.Client) {
	rat.currClient = client

	for _, mod := range rat.All() {
		if mod != rat {
			mod.SetClient(client)
		}
	}
}

// Close terminates the remote administration tool.
func (rat *rat) Close() error {

	if rat.abort != nil {
		close(rat.abort)
	}

	if rat.listener != nil {
		rat.listener.Close()
	}

	rat.sockets.Close()

	rat.waitGroup.Wait()
	rat.LogStorage()
	rat.Log("The server has exited.")
	return nil
}

func (rat *rat) listen() {
	defer rat.waitGroup.Done()
	defer rat.Store("The listen routine has exited.")

	for {
		conn, err := rat.listener.Accept()
		if err != nil {
			select {
			case <-rat.abort:
			default:
				rat.Store(err)
			}

			return
		}

		// A new client has connected to the server.
		client := ratnet.NewClient(conn)
		rat.Store(fmt.Sprintf("A new client [%v] has connected.", client.ID()))
		socket := rat.sockets.Add(client)

		rat.waitGroup.Add(1)
		go rat.transfer(socket)
	}
}

func (rat *rat) transfer(socket *socket) {
	defer rat.waitGroup.Done()
	defer rat.Store(
		fmt.Sprintf("The transfer routine of client [%v] has exited.", socket.client.ID()))

	client := socket.client
	for {
		pkg, err := client.RecvPacket()
		if err != nil {
			select {
			case <-socket.abort:
			default:
				rat.sockets.Del(client.ID())
				rat.Store(err)
			}

			return
		}

		mod := rat.ByPacket(pkg.Type)
		if mod == nil {
			rat.Store(
				fmt.Errorf("The client [%v] has received a packet with invalid type: %v",
					client.ID(), pkg.Type))
			continue
		}

		err = mod.Respond(client, pkg)
		if err != nil {
			if errors.Is(err, io.EOF) {
				rat.sockets.Del(client.ID())
				return
			}

			rat.Store(err)
		}
	}
}

func (rat *rat) getSocket(id interface{}) (*socket, error) {
	cid, err := ratnet.ToID(id)
	if err != nil {
		return nil, err
	}

	socket := rat.sockets.Get(cid)
	if socket == nil {
		return nil, fmt.Errorf("Invalid client ID: %v", id)
	}

	return socket, nil
}

// ------ Command Handlers ------

func (rat *rat) switchClient(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Null argument")
	}

	socket, err := rat.getSocket(args[0])
	if err != nil {
		return err
	}

	rat.SetClient(socket.client)
	rat.Log(fmt.Sprintf("The current client has changed to [%v].", rat.currClient.ID()))
	return nil
}

func (rat *rat) terminateClient(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Null argument")
	}

	socket, err := rat.getSocket(args[0])
	if err != nil {
		return err
	}

	return rat.onDisconnect(socket.client, nil)
}

// ------ Packet Handlers ------

func (rat *rat) onDisconnect(client ratnet.Client, pkg *ratnet.Packet) error {
	panic.Assert(client != nil, "Null argument.")

	id := client.ID()
	if rat.sockets.Del(id) {
		rat.Store(fmt.Sprintf("The client [%v] has disconnected.", id))
		return nil
	}

	return fmt.Errorf("Invalid client ID: %v", id)
}
