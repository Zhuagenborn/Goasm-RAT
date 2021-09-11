package mod

import (
	"fmt"

	"server/internal/net"
	"server/internal/utility/panic"
)

// Dispatcher provides module management.
type Dispatcher interface {

	// Register registers a module.
	Register(Mod) error

	// ByID finds a module by its ID.
	ByID(ID) Mod

	// ByCmd finds a module by a supported command.
	ByCmd(string) Mod

	// ByPacket finds a module by a packet type.
	ByPacket(net.PktType) Mod

	// All gets all registered modules.
	All() []Mod
}

type modList map[ID]Mod

type cmdList map[string]Mod

type pkgList map[net.PktType]Mod

type dispatcher struct {
	mods modList
	cmds cmdList
	pkgs pkgList
}

// NewDispatcher creates a new dispatcher.
func NewDispatcher() Dispatcher {
	return &dispatcher{
		mods: make(modList),
		cmds: make(cmdList),
		pkgs: make(pkgList),
	}
}

func (dp *dispatcher) Register(mod Mod) error {
	panic.Assert(mod != nil, "Null module.")

	id := mod.ID()
	if dp.ByID(id) != nil {
		return fmt.Errorf("Conflicting ID: %d", id)
	}

	dp.mods[id] = mod

	for _, s := range mod.Cmds() {
		if dp.ByCmd(s) != nil {
			return fmt.Errorf("Conflicting command: %s", s)
		}

		dp.cmds[s] = mod
	}

	for _, t := range mod.Packets() {
		if dp.ByPacket(t) != nil {
			return fmt.Errorf("Conflicting packet type: %v", t)
		}

		dp.pkgs[t] = mod
	}

	return nil
}

func (dp *dispatcher) ByID(id ID) Mod {
	if mod, ok := dp.mods[id]; ok {
		return mod
	}

	return nil
}

func (dp *dispatcher) ByCmd(cmd string) Mod {
	if mod, ok := dp.cmds[cmd]; ok {
		return mod
	}

	return nil
}

func (dp *dispatcher) ByPacket(pkg net.PktType) Mod {
	if mod, ok := dp.pkgs[pkg]; ok {
		return mod
	}

	return nil
}

func (dp *dispatcher) All() []Mod {
	mods := make([]Mod, 0, len(dp.mods))
	for _, mod := range dp.mods {
		mods = append(mods, mod)
	}

	return mods
}
