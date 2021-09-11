package dummy

import (
	"fmt"

	proto "server/internal/mod"
	"server/internal/net"
)

type mod struct {
	id   proto.ID
	cmds []string
	pkgs []net.PktType
}

// NewMod creates a dummy module for tests.
func NewMod(id proto.ID, cmds []string, pkgs []net.PktType) proto.Mod {
	return &mod{id, cmds, pkgs}
}

func (*mod) Exec(cmd string, args []string) error {
	return nil
}

func (mod *mod) Cmds() []string {
	return mod.cmds
}

func (*mod) Respond(client net.Client, pkg *net.Packet) error {
	return nil
}

func (mod *mod) Packets() []net.PktType {
	return mod.pkgs
}

func (mod *mod) ID() proto.ID {
	return mod.id
}

func (*mod) Name() string {
	return "DUMMY"
}

func (*mod) SetClient(net.Client) {
}

func (*mod) Close() error {
	return nil
}

func (mod *mod) String() string {
	return fmt.Sprintf("%d-%s", mod.ID(), mod.Name())
}
