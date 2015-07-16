package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/sqs"
)

//QuestionQName : name of the questions SQS Queue
const QuestionQName = "demo-questions"

//AnswerQName : name of the answers SQS Queue
const AnswerQName = "demo-answers"

//SQS instance of the sqs var from goamz
var SQS *sqs.SQS

//Question : struct that holds the data for the math problem.
type Question struct {
	Num1 int
	Num2 int
}

var sem = make(chan int, 250)

func main() {

}

func Serve(queue chan []sqs.Message, sqsQ *sqs.Queue) {
	log.Println("In Serve")
	for q := range queue {
		log.Println("In range")
		sem <- 1
		go func() {
			q := q
			addToQuestionQ(q, *sqsQ)
			<-sem
		}()
	}
	log.Println("Done")
}

func makeRun(public string, secret string, maxworkers int) error {

	auth := aws.Auth{AccessKey: public, SecretKey: secret}
	region := aws.Region{}
	region.Name = "us-west-2"
	region.SQSEndpoint = "http://sqs.us-west-2.amazonaws.com"
	SQS = sqs.New(auth, region)
	if SQS == nil {
		return fmt.Errorf("Can't get sqs reference for %v %v", auth, region)
	}
	questionq, getErr := SQS.GetQueue(QuestionQName)
	if getErr != nil {
		fmt.Println(QuestionQName)
		return getErr
	}

	msgSlice := []sqs.Message{}
	var queue chan []sqs.Message
	log.Println("Before Serve")
	Serve(queue, questionq)
	log.Println("After Serve")

	for i := 0; i < 100000; i++ {
		num1 := rand.Intn(9) + 1
		num2 := rand.Intn(9 + 1)
		q := Question{num1, num2}
		jsonQ, _ := json.Marshal(&q)
		msg := sqs.Message{}
		body := base64.StdEncoding.EncodeToString(jsonQ)
		msg.Body = body
		msgSlice = append(msgSlice, msg)
		if len(msgSlice) == 10 {
			//the slice is ready, time to process it
			//I guess we pass it to the Serve function
			log.Println("Send to Channel")
			queue <- msgSlice
			msgSlice = []sqs.Message{}
		}
	}

	return nil
}

func addToQuestionQ(msgList []sqs.Message, sqsQ sqs.Queue) {
	_, respErr := sqsQ.SendMessageBatch(msgList)
	if respErr != nil {
		log.Println("ERROR:", respErr)
	}
}
