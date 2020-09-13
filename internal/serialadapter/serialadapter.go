package serialadapter

import (
	"io"
	"log"
	"strings"

	"github.com/jacobsa/go-serial/serial"
)

// Command struct
type Command struct {
	Motor   string `json:"motorName"`
	Degrees string `json:"rotationDegrees"`
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
	commandAsStringArray := []string{command.Motor, command.Degrees}
	commandAsString := strings.Join(commandAsStringArray[:], ",")
	s.Port.Write([]byte(commandAsString))
	log.Printf("serialadapter wrote message to serial port: %s", commandAsString)
}
