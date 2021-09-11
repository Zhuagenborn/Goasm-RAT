package net

// PktType is used to specify a packet's type.
type PktType int32

const (
	// Unknow is the default type of a packet, it means nothing.
	Unknow PktType = 0
)

// Header is the structure of a packet's header.
type Header struct {
	Type     PktType
	DataSize int32
}
