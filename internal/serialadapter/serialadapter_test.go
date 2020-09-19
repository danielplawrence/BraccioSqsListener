package serialadapter

import (
	"io"
	"testing"

	"github.com/jacobsa/go-serial/serial"
	"github.com/stretchr/testify/mock"
)

type MockReadWriteCloser struct {
	mock.Mock
}

func (*MockReadWriteCloser) Close() error {
	return nil
}
func (*MockReadWriteCloser) Read(p []byte) (n int, err error) {
	return 0, nil
}
func (r *MockReadWriteCloser) Write(p []byte) (n int, err error) {
	r.Called(p)
	return 0, nil
}

type MockSerialClient struct {
	mock.Mock
	MockedPort io.ReadWriteCloser
}

func (c MockSerialClient) Open(options serial.OpenOptions) (io.ReadWriteCloser, error) {
	c.Called(options)
	return c.MockedPort, nil
}

func TestNew(t *testing.T) {
	options := serial.OpenOptions{
		PortName:        "/dev/cu.usbmodem143101",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	port := new(MockReadWriteCloser)
	client := MockSerialClient{
		MockedPort: port,
	}
	client.On("Open", options).Return(&MockReadWriteCloser{})
	New(client, options)
}

func TestSend(t *testing.T) {
	options := serial.OpenOptions{
		PortName:        "/dev/cu.usbmodem143101",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	command := Command{
		Motor:   "1",
		Degrees: "100",
	}
	expectedBytes := []byte("1,100")
	port := new(MockReadWriteCloser)
	client := MockSerialClient{
		MockedPort: port,
	}
	client.On("Open", options).Return(port)
	port.On("Write", expectedBytes).Return(1)
	adapter := New(client, options)
	adapter.SendCommand(&command)
}
