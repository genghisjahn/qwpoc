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

var bufferCount = 100 //Playing with this get's different speep through puts.
//The mbAir needs it around 50 or it starts erroring out.
//The mbPro can handle higher values, but diminishing returns.

//But, even set at 50, the mbAir can put 100,000 items in the queu in  48 seconds.

var sem = make(chan bool, bufferCount)

func main() {
}

func doRun(public string, secret string, maxworkers int) error {

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
	msgSlice := make([]sqs.Message, 0, 10) //A slice that holds up to 10 messages
	msgAll := [][]sqs.Message{}

	for i := 0; i < 10000; i++ {
		num1 := rand.Intn(9) + 1
		num2 := rand.Intn(9 + 1)
		q := Question{num1, num2}
		jsonQ, _ := json.Marshal(&q)
		msg := sqs.Message{Body: base64.StdEncoding.EncodeToString(jsonQ)}
		msgSlice = append(msgSlice, msg)
		if len(msgSlice) == 10 {
			msgAll = append(msgAll, msgSlice)
			msgSlice = []sqs.Message{}
		}
	}
	for _, s := range msgAll {
		s := s //It's idomatic go I swear! http://golang.org/doc/effective_go.html#channels
		sem <- true
		go func(sl10 []sqs.Message) {
			addToQuestionQ(sl10, questionq)
			defer func() { <-sem }()
		}(s)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return nil
}

func addToQuestionQ(msgList []sqs.Message, sqsQ *sqs.Queue) {
	_, respErr := sqsQ.SendMessageBatch(msgList)
	if respErr != nil {
		log.Println("ERROR:", respErr)
	}
}
