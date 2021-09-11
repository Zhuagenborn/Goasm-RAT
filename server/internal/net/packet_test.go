package net_test

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"

	. "server/internal/net"
)

func TestPacket_Write(t *testing.T) {
	cases := []struct {
		segs   [][]byte
		expect []byte
	}{
		{nil, nil},
		{[][]byte{{1, 2, 3}}, []byte{1, 2, 3}},
		{[][]byte{{1, 2}, {3}}, []byte{1, 2, 3}},
		{[][]byte{{1}, {2}, {3}}, []byte{1, 2, 3}},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			pkg := NewPacket()
			// Write all segments to the packet.
			for _, seg := range c.segs {
				n, err := pkg.Write(seg)
				if n != len(seg) || err != nil {
					t.Fatalf("Expect <%d>, but get <%d>; Error: <%v>.", len(seg), n, err)
				}
			}

			// Check all data.
			if !bytes.Equal(c.expect, pkg.Data) {
				t.Errorf("Expect <%v>, but get <%v>.", c.expect, pkg.Data)
			}
		})
	}
}

func TestPacket_Read(t *testing.T) {

	type seg struct {
		size   int32
		expect []byte
	}

	cases := []struct {
		all  []byte
		segs []seg
	}{
		{nil, nil},
		{[]byte{1, 2}, []seg{{1, []byte{1}}, {1, []byte{2}}}},
		{[]byte{1, 2, 3}, []seg{{2, []byte{1, 2}}, {1, []byte{3}}}},
		{[]byte{1, 2, 3}, []seg{{3, []byte{1, 2, 3}}}},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			pkg := NewPacket()
			pkg.Data = c.all
			pkg.DataSize = int32(len(c.all))

			// Read each segment from the packet.
			for _, seg := range c.segs {
				buf := make([]byte, seg.size)
				n, _ := pkg.Read(buf)
				if n != int(seg.size) || !bytes.Equal(seg.expect, buf) {
					t.Fatalf("Expect <%v>, but get <%v>.", seg.expect, buf)
				}
			}

			// The packet should be empty and check the `io.EOF`.
			if n, err := pkg.Read(make([]byte, 1)); n != 0 || err != io.EOF {
				t.Errorf("Expect <%d> and <%v>, but get <%d> and <%v>.", 0, io.EOF, n, err)
			}
		})
	}
}

func TestPacket_ToBinary(t *testing.T) {
	cases := []struct {
		pkg    Packet
		expect []byte
	}{
		{
			Packet{Header{PktType(1), 4}, []byte{1, 2, 3, 4}},
			[]byte{1, 0, 0, 0, 4, 0, 0, 0, 1, 2, 3, 4},
		},
		{
			Packet{Header{PktType(2), 1}, []byte{1}},
			[]byte{2, 0, 0, 0, 1, 0, 0, 0, 1},
		},
		{
			Packet{Header{PktType(3), 0}, nil},
			[]byte{3, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			raw := c.pkg.ToBinary()
			if !bytes.Equal(c.expect, raw) {
				t.Errorf("Expect <%v>, but get <%v>.", c.expect, raw)
			}
		})
	}
}

func TestPacket_FromBinary(t *testing.T) {
	cases := []struct {
		bytes  []byte
		expect *Packet
	}{
		{
			[]byte{1, 0, 0, 0, 4, 0, 0, 0, 1, 2, 3, 4},
			&Packet{Header{PktType(1), 4}, []byte{1, 2, 3, 4}},
		},
		{
			[]byte{2, 0, 0, 0, 1, 0, 0, 0, 1},
			&Packet{Header{PktType(2), 1}, []byte{1}},
		},
		{
			[]byte{3, 0, 0, 0, 0, 0, 0, 0},
			&Packet{Header{PktType(3), 0}, nil},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			pkg := NewPacket()
			if err := pkg.FromBinary(c.bytes); err != nil || !reflect.DeepEqual(c.expect, pkg) {
				t.Errorf("Expect <%v>, but get <%v>; Error: <%v>.", c.expect, pkg, err)
			}
		})
	}

}
