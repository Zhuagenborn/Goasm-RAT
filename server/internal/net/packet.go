package net

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

// Packet is the structure of a network packet.
type Packet struct {
	Header
	Data []byte
}

// NewPacket creates a new empty packet.
func NewPacket() *Packet {
	return &Packet{}
}

// Write writes data to a packet.
func (pkg *Packet) Write(data []byte) (int, error) {
	pkg.Data = append(pkg.Data, data...)
	pkg.DataSize += int32(len(data))
	return len(data), nil
}

// Read reads data from a packet.
func (pkg *Packet) Read(buffer []byte) (int, error) {
	if len(pkg.Data) == 0 {
		return 0, io.EOF
	}

	if cap(buffer) >= len(pkg.Data) {
		// Read all data.
		copy(buffer, pkg.Data)
		pkg.Data = nil
		return len(buffer), io.EOF
	}

	// Read a part of data.
	copy(buffer, pkg.Data[:cap(buffer)])
	pkg.Data = pkg.Data[cap(buffer):]
	return len(buffer), nil
}

// Close release the data.
func (pkg *Packet) Close() error {
	pkg.Data = nil
	pkg.DataSize = 0
	return nil
}

// ToBinary converts a packet to binary data.
func (pkg *Packet) ToBinary() []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.LittleEndian, pkg.Type)
	binary.Write(buffer, binary.LittleEndian, pkg.DataSize)
	binary.Write(buffer, binary.LittleEndian, pkg.Data)
	return buffer.Bytes()
}

// FromBinary creates a packet from binary data.
func (pkg *Packet) FromBinary(data []byte) error {
	headerSize := int(unsafe.Sizeof(Header{}))
	if len(data) < headerSize {
		return fmt.Errorf("Not enough data")
	}

	reader := bytes.NewReader(data)
	binary.Read(reader, binary.LittleEndian, &pkg.Type)
	binary.Read(reader, binary.LittleEndian, &pkg.DataSize)
	if pkg.DataSize > 0 {
		pkg.Data = make([]byte, pkg.DataSize)
		copy(pkg.Data, data[headerSize:])
	}

	return nil
}
