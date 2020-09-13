package sqslistener

import (
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// HandlerFunc is used to define the Handler that is run on for each message
type HandlerFunc func(msg *sqs.Message) error

// HandleMessage wraps a function for handling sqs messages
func (f HandlerFunc) HandleMessage(msg *sqs.Message) error {
	return f(msg)
}

// Handler interface
type Handler interface {
	HandleMessage(msg *sqs.Message) error
}

// Listener struct
type Listener struct {
	Config    *Config
	SqsClient sqsiface.SQSAPI
}

// Config struct
type Config struct {
	MaxNumberOfMessage int64
	QueueURL           string
	WaitTimeSecond     int64
	MaxPolls           int
}

// New listener
func New(client sqsiface.SQSAPI, config *Config) *Listener {
	return &Listener{
		Config:    config,
		SqsClient: client,
	}
}

// Start the listener
func (listener *Listener) Start(handler Handler) {
	log.Print("Polling for messages")
	for i := 0; i < listener.Config.MaxPolls; i++ {
		log.Print(fmt.Sprintf("Reading messages from queue %s", listener.Config.QueueURL))
		params := &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(listener.Config.QueueURL),
			MaxNumberOfMessages: aws.Int64(listener.Config.MaxNumberOfMessage),
			AttributeNames: []*string{
				aws.String("All"),
			},
			WaitTimeSeconds: aws.Int64(listener.Config.WaitTimeSecond),
		}
		resp, err := listener.SqsClient.ReceiveMessage(params)
		if err != nil {
			log.Println(err)
		}
		if len(resp.Messages) > 0 {
			listener.run(handler, resp.Messages)
		}
	}
}

func (listener *Listener) run(handler Handler, messages []*sqs.Message) {
	numMessages := len(messages)
	log.Print(fmt.Sprintf("Received %d messages", numMessages))
	var wg sync.WaitGroup
	wg.Add(numMessages)
	for i := range messages {
		go func(m *sqs.Message) {
			// launch goroutine
			defer wg.Done()
			err := listener.handle(handler, m)
			if err != nil {
				log.Print(err.Error())
			}
		}(messages[i])
	}

	wg.Wait()
}

func (listener *Listener) handle(handler Handler, message *sqs.Message) error {
	log.Print(fmt.Sprintf("Running handler for message %s", message))
	var err error

	err = handler.HandleMessage(message)
	if err != nil {
		return err
	}

	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(listener.Config.QueueURL),
		ReceiptHandle: message.ReceiptHandle,
	}
	_, err = listener.SqsClient.DeleteMessage(params)

	if err != nil {
		return err
	}

	log.Print(fmt.Sprintf("worker: deleted message from queue: %s", aws.StringValue(message.ReceiptHandle)))

	return nil
}
