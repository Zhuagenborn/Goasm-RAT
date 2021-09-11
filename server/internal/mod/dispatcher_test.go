package mod_test

import (
	"fmt"
	"os"
	"testing"

	"server/internal/mod"
	. "server/internal/mod"
	"server/internal/net"
	"server/internal/test/dummy"
	"server/internal/utility/panic"
)

var dp Dispatcher

const INVALID_ID = -1

func TestMain(m *testing.M) {
	mods := []Mod{
		dummy.NewMod(1, []string{"1-1"}, []net.PktType{10}),
		dummy.NewMod(2, []string{"2-1", "2-2"}, []net.PktType{21, 22}),
		dummy.NewMod(3, []string{"3-1"}, []net.PktType{}),
		dummy.NewMod(4, []string{}, []net.PktType{}),
	}

	dp = NewDispatcher()
	for _, mod := range mods {
		if err := dp.Register(mod); err != nil {
			panic.Assert(false, fmt.Sprintf("Fail to register <%s>.", mod))
		}
	}

	ret := m.Run()
	os.Exit(ret)
}

func TestDispatcher_Register(t *testing.T) {
	cases := []struct {
		mod     Mod
		expPass bool
	}{
		{dummy.NewMod(1, []string{"1"}, []net.PktType{1}), true},
		{dummy.NewMod(2, []string{"2"}, []net.PktType{2}), true},
		{dummy.NewMod(3, []string{"1"}, []net.PktType{3}), false},
		{dummy.NewMod(4, []string{"4"}, []net.PktType{2}), false},
		{dummy.NewMod(4, []string{"5"}, []net.PktType{5}), false},
	}

	dp := NewDispatcher()
	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if err := dp.Register(c.mod); c.expPass && err != nil {
				t.Errorf("Expect a success, but get <%v>.", err)
			} else if !c.expPass && err == nil {
				t.Errorf("Expect an error, but get <%v>.", nil)
			}
		})
	}
}

func TestDispatcher_ByID(t *testing.T) {
	cases := []struct {
		id      mod.ID
		expPass bool
	}{
		{1, true},
		{2, true},
		{6, false},
		{7, false},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			mod := dp.ByID(c.id)
			if c.expPass && mod == nil {
				t.Errorf("Expect the Module <%d>, but get <%v>.", c.id, nil)
			} else if !c.expPass && mod != nil {
				t.Errorf("Expect <%v>, but get <%v>.", nil, mod)
			}
		})
	}
}

func TestDispatcher_ByCmd(t *testing.T) {
	cases := []struct {
		cmd   string
		expID mod.ID
	}{
		{"1-1", 1},
		{"2-1", 2},
		{"2-3", INVALID_ID},
		{"3-3", INVALID_ID},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if mod := dp.ByCmd(c.cmd); !checkMod(c.expID, mod) {
				t.Errorf("Expect the Module <%d>, but get <%v>.", c.expID, mod)
			}
		})
	}
}

func checkMod(expID mod.ID, mod mod.Mod) bool {
	if expID == INVALID_ID {
		if mod == nil {
			return true
		} else {
			return false
		}
	} else {
		if mod == nil {
			return false
		} else {
			return mod.ID() == expID
		}
	}
}
