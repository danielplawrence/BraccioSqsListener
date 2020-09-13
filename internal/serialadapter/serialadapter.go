package serialadapter

import (
	"io"
	"log"
	"strconv"

	"github.com/jacobsa/go-serial/serial"
)

// Command struct
type Command struct {
	Motor    string `json:"motorNumber"`
	Degrees  string `json:"rotationDegrees"`
	Rotation string `json:"direction"`
}

// SerialClient injector
type SerialClient interface {
	Open(options serial.OpenOptions) (io.ReadWriteCloser, error)
}

// SerialAdapter struct
type SerialAdapter struct {
	Port io.ReadWriteCloser
}

// New serial adapter
func New(client SerialClient, serialOptions serial.OpenOptions) *SerialAdapter {
	// Open the port.
	port, err := client.Open(serialOptions)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	return &SerialAdapter{
		Port: port,
	}
}

// SendCommand sends a command to the serial port
func (s *SerialAdapter) SendCommand(command *Command) {
	// Our serial command should look like
	// motorNumber(int) Degrees(int) Rotation(-1 or +1)
	motorNumber, _ := strconv.Atoi(command.Motor)
	degrees, _ := strconv.Atoi(command.Degrees)
	m := make(map[string]int)
	m["clockwise"] = 1
	m["anticlockwise"] = -1
	rotation := m[command.Rotation]
	var byteArray []byte
	byteArray = append(byteArray, byte(motorNumber))
	byteArray = append(byteArray, byte(degrees))
	byteArray = append(byteArray, byte(rotation))
	s.Port.Write(byteArray)
	log.Printf("serialadapter wrote message to serial port:")
	log.Print(byteArray)
}
