package sqslistener

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type sqsEvent struct {
	Foo string `json:"foo"`
	Qux string `json:"qux"`
}

type mockSQS struct {
	mock.Mock
	sqsiface.SQSAPI
	messages map[string][]*sqs.Message
}

func (m *mockSQS) SendMessage(in *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	m.messages[*in.QueueUrl] = append(m.messages[*in.QueueUrl], &sqs.Message{
		Body: in.MessageBody,
	})
	return &sqs.SendMessageOutput{}, nil
}
func (m *mockSQS) ReceiveMessage(in *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	if len(m.messages[*in.QueueUrl]) == 0 {
		return &sqs.ReceiveMessageOutput{}, nil
	}
	response := m.messages[*in.QueueUrl][0:1]
	m.messages[*in.QueueUrl] = m.messages[*in.QueueUrl][1:]
	return &sqs.ReceiveMessageOutput{
		Messages: response,
	}, nil
}
func (m *mockSQS) GetQueueUrl(urlInput *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	url := "http://someurl.com"
	return &sqs.GetQueueUrlOutput{QueueUrl: &url}, nil
}
func (m *mockSQS) DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	m.Called(input)
	return &sqs.DeleteMessageOutput{}, nil
}

type mockedHandler struct {
	mock.Mock
}

func (mh *mockedHandler) HandleMessage(message *sqs.Message) {
	mh.Called(message)
}

func TestNew(t *testing.T) {
	const maxNumberOfMessages = 1984
	const waitTimeSecond = 1337
	client := &mockSQS{
		messages: map[string][]*sqs.Message{},
	}
	config := &Config{
		MaxNumberOfMessage: maxNumberOfMessages,
		QueueURL:           "http://someurl.com",
		WaitTimeSecond:     waitTimeSecond,
		MaxPolls:           1,
	}
	worker := New(client, config)
	assert.Equal(t, config, worker.Config, "Config was set properly")
}

func TestStart(t *testing.T) {
	const maxNumberOfMessages = 1984
	const waitTimeSecond = 1337
	queueURL := "http://someurl.com"
	client := &mockSQS{
		messages: map[string][]*sqs.Message{},
	}
	config := &Config{
		MaxNumberOfMessage: maxNumberOfMessages,
		QueueURL:           "http://someurl.com",
		WaitTimeSecond:     waitTimeSecond,
		MaxPolls:           1,
	}
	worker := New(client, config)
	handler := new(mockedHandler)
	handlerFunc := HandlerFunc(func(msg *sqs.Message) (err error) {
		handler.HandleMessage(msg)
		return
	})
	expectedString := "Hello world"
	expectedMessage := &sqs.Message{
		Body: &expectedString,
	}
	expectedDeleteInput := &sqs.DeleteMessageInput{QueueUrl: &queueURL}

	t.Run("the listener successfully processes a message", func(t *testing.T) {
		client.SendMessage(&sqs.SendMessageInput{
			MessageBody: aws.String(expectedString),
			QueueUrl:    &queueURL,
		})
		handler.On("HandleMessage", expectedMessage).Return().Once()
		client.On("DeleteMessage").On("DeleteMessage", expectedDeleteInput).Return()
		worker.Start(handlerFunc)
		handler.AssertExpectations(t)
	})
}
