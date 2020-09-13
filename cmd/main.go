package main

import (
	"braccio/listener/internal/serialadapter"
	"braccio/listener/internal/sqslistener"
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/jacobsa/go-serial/serial"
)

type RealSerialClient struct {
}

func (c RealSerialClient) Open(options serial.OpenOptions) (io.ReadWriteCloser, error) {
	return serial.Open(options)
}

func main() {
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sqsClient := sqs.New(awsSession)
	listenerConfig := &sqslistener.Config{
		QueueURL:           "https://sqs.us-east-1.amazonaws.com/066043029358/BraccioSkillQueue",
		MaxNumberOfMessage: 1,
		WaitTimeSecond:     5,
		MaxPolls:           1000000000,
	}
	serialOptions := serial.OpenOptions{
		PortName:        "/dev/tty.Bluetooth-Incoming-Port",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	listener := sqslistener.New(sqsClient, listenerConfig)
	serialAdapter := serialadapter.New(new(RealSerialClient), serialOptions)
	// set up the handler func
	messageHandler := sqslistener.HandlerFunc(func(msg *sqs.Message) error {
		// Get the JSON and convert to Command
		command := &serialadapter.Command{}
		json.Unmarshal([]byte(aws.StringValue(msg.Body)), command)
		// Send to serial port
		serialAdapter.SendCommand(command)
		return nil
	})
	// start the listener
	listener.Start(messageHandler)
}
