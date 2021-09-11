package rat

import (
	"fmt"
	"testing"

	"server/internal/net"
)

const (
	Add = 0
	Get = 1
	Del = 2
)

type action struct {
	cmd     int
	expFail bool
}

func TestSocketList(t *testing.T) {
	cases := []struct {
		acts []action
	}{
		{[]action{{Get, true}, {Add, false}, {Get, false}, {Del, false}, {Get, true}}},
		{[]action{{Del, true}, {Get, true}, {Add, false}, {Del, false}, {Get, true}}},
		{[]action{{Add, false}, {Add, false}, {Get, false}, {Del, false}, {Get, true}}},
		{[]action{{Add, false}, {Del, false}, {Del, true}}},
	}

	client := net.NewClient(nil)
	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			list := newSocketList()
			for j, act := range c.acts {
				var ret interface{} = nil
				switch act.cmd {
				case Add:
					ret = list.Add(client)
				case Del:
					ret = list.Del(client.ID())
				case Get:
					ret = list.Get(client.ID())
				default:
					t.SkipNow()
				}

				fail := false
				switch val := ret.(type) {
				case bool:
					fail = !val
				case *socket:
					fail = val == nil
				default:
					t.Fatal("The retuen type is error.")
				}

				if act.expFail && !fail {
					t.Fatalf("Expect a failure in Action <%d>, but get <%v>.", j, ret)
				} else if !act.expFail && fail {
					t.Fatalf("Expect a success in Action <%d>, but get <%v>.", j, ret)
				}
			}
		})
	}
}
