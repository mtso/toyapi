package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func getMessages(svc *sqs.SQS, qUrl string, done chan *sqs.Message) {
	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            &qUrl,
		MaxNumberOfMessages: aws.Int64(2),
		VisibilityTimeout:   aws.Int64(36000), // 10 hours
		WaitTimeSeconds:     aws.Int64(0),
	})

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	if len(result.Messages) == 0 {
		fmt.Println("Received no messages")
		return
	}

	_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &qUrl,
		ReceiptHandle: result.Messages[0].ReceiptHandle,
	})

	if err != nil {
		fmt.Println("Delete Error", err)
		return
	}

	for _, msg := range result.Messages {
		done <- msg
	}
}

func main() {
	sess := session.Must(session.NewSession())
	svc := sqs.New(sess)
	qUrl := os.Getenv("QURL")

	t := time.NewTicker(time.Second * 2)
	recv := make(chan *sqs.Message)

	for {
		select {
		case m := <-recv:
			fmt.Println(m)

		case <-t.C:
			getMessages(svc, qUrl, recv)
		}
	}
}
