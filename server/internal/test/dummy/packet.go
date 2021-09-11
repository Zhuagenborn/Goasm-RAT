package dummy

import (
	"math"
	"math/rand"
	"time"

	"server/internal/net"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomPacket creates a random packet.
func RandomPacket() *net.Packet {
	pkg := net.NewPacket()
	pkg.Type = net.PktType(rand.Intn(math.MaxUint8))
	pkg.DataSize = int32(rand.Intn(10))
	pkg.Data = make([]byte, pkg.DataSize)
	for i := 0; i != int(pkg.DataSize); i++ {
		pkg.Data[i] = byte(rand.Intn(math.MaxUint8))
	}

	return pkg
}
