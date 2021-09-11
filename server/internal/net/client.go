package net

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unsafe"
)

// ID is used to uniquely specify a client.
type ID int

// ToID converts an integer number or a string into a client's ID.
func ToID(num interface{}) (ID, error) {
	switch num.(type) {
	case string:
		id, err := strconv.Atoi(num.(string))
		return ID(id), err
	case int:
		return ID(num.(int)), nil
	case ID:
		return num.(ID), nil
	default:
		return 0, fmt.Errorf("Invalid type: %s", reflect.TypeOf(num))
	}
}

var newID ID = 0

// Client is the interface of the client.
type Client interface {
	io.Closer

	fmt.Stringer

	// ID returns the client's ID.
	ID() ID

	// RecvPacket receives a packet from the client.
	RecvPacket() (*Packet, error)

	// SendPacket sends a packet to the client.
	SendPacket(*Packet) error
}

type client struct {
	id ID
	io.ReadWriteCloser
}

// NewClient creates a new client.
func NewClient(conn io.ReadWriteCloser) Client {
	newID++
	return &client{
		id:              newID,
		ReadWriteCloser: conn,
	}
}

func (client *client) RecvPacket() (*Packet, error) {
	pkg := NewPacket()
	// Receive the header.
	header, err := client.recvData(int(unsafe.Sizeof(Header{})))
	if err != nil && err != io.EOF {
		return nil, err
	}

	reader := bytes.NewReader(header)
	binary.Read(reader, binary.LittleEndian, &pkg.Type)
	binary.Read(reader, binary.LittleEndian, &pkg.DataSize)

	// Receive the body.
	pkg.Data, err = client.recvData(int(pkg.DataSize))
	if err != nil && err != io.EOF {
		return nil, err
	}

	return pkg, nil
}

func (client *client) SendPacket(pkg *Packet) error {
	return client.sendData(pkg.ToBinary())
}

// String converts a client into a string.
func (client *client) String() string {
	return fmt.Sprintf("%d", client.id)
}

func (client *client) ID() ID {
	return client.id
}

func (client *client) Close() error {
	return nil
}

// recvData receives raw data from the client.
func (client *client) recvData(size int) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	temp := make([]byte, size)
	for received := 0; received < size; {
		curr, err := client.Read(temp)
		if err != nil && err != io.EOF {
			return buffer.Bytes(), err
		}

		buffer.Write(temp[:curr])
		received += curr
	}

	return buffer.Bytes(), nil
}

// sendData sends raw data to the client.
func (client *client) sendData(data []byte) error {
	for sent := 0; sent < len(data); {
		curr, err := client.Write(data[sent:])
		if err != nil {
			return err
		}

		sent += curr
	}

	return nil
}
