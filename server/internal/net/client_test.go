package net_test

import (
	"fmt"
	"io"
	"reflect"
	"testing"

	. "server/internal/net"
	"server/internal/test/dummy"
)

func TestToID(t *testing.T) {
	cases := []struct {
		num    interface{}
		expect ID
	}{
		{1, ID(1)},
		{"2", ID(2)},
		{ID(3), ID(3)},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if id, err := ToID(c.num); id != c.expect || err != nil {
				t.Errorf("Expect <%v>, but get <%v>; Error: <%v>.", c.expect, id, err)
			}
		})
	}
}

// Use `Packet` as the stub for a remote client.

func TestClient_RecvPacket(t *testing.T) {
	remote := NewPacket()
	client := NewClient(remote)
	for i := 0; i != 5; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			remote.Close()
			expect := dummy.RandomPacket()
			remote.Write(expect.ToBinary())
			if recv, err := client.RecvPacket(); (err != nil && err != io.EOF) || !reflect.DeepEqual(expect, recv) {
				t.Errorf("Expect <%v>, but get <%v>; Error: <%v>.", expect, recv, err)
			}
		})
	}
}

func TestClient_SendPacket(t *testing.T) {
	remote := NewPacket()
	client := NewClient(remote)
	for i := 0; i != 5; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			remote.Close()
			pkg := dummy.RandomPacket()
			if err := client.SendPacket(pkg); err != nil {
				t.Fatalf("Error: <%v>.", err)
			}

			bytes := pkg.ToBinary()
			if !reflect.DeepEqual(bytes, remote.Data) {
				t.Errorf("Expect <%v>, but get <%v>.", bytes, remote.Data)
			}
		})
	}
}
