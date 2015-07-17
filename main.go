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

//ProblemQName : name of the problems SQS Queue
const ProblemQName = "demo-problems"

//ProblemSharedQName : name of the problem shard SQS Queue
const ProblemSharedQName = "demo-problem-shard"

//AnswerQName : name of the answers SQS Queue
const AnswerQName = "demo-answers"

//SQS instance of the sqs var from goamz
var SQS *sqs.SQS

//Problem : struct that holds the data for the math problem.
type Problem struct {
	Num1 int
	Num2 int
}

var bufferCount = 200 //Playing with this get's different speep through puts.

var sem = make(chan bool, bufferCount)

func main() {
}

func doRun(public string, secret string, maxworkers int) error {

	auth := aws.Auth{AccessKey: public, SecretKey: secret}
	region := aws.Region{}
	region.Name = "us-east-1"
	region.SQSEndpoint = "http://sqs.us-east-1.amazonaws.com"
	SQS = sqs.New(auth, region)
	if SQS == nil {
		return fmt.Errorf("Can't get sqs reference for %v %v", auth, region)
	}
	problemq, getErr := SQS.GetQueue(ProblemQName)
	if getErr != nil {
		fmt.Println(ProblemQName)
		return getErr
	}
	msgSlice := make([]sqs.Message, 0, 10) //A slice that holds up to 10 messages
	msgAll := [][]sqs.Message{}

	for i := 0; i < 1000; i++ {
		num1 := rand.Intn(9) + 1
		num2 := rand.Intn(9 + 1)
		p := Problem{num1, num2}
		jsonP, _ := json.Marshal(&p)
		msg := sqs.Message{Body: base64.StdEncoding.EncodeToString(jsonP)}
		msgSlice = append(msgSlice, msg)
		if len(msgSlice) == 10 {
			msgAll = append(msgAll, msgSlice)
			msgSlice = []sqs.Message{}
		}
	}
	for _, s := range msgAll {
		sem <- true
		go func(sl10 []sqs.Message) {
			addToProblemQ(sl10, *problemq)
			defer func() { <-sem }()
		}(s)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return nil
}

func addToProblemQ(msgList []sqs.Message, sqsQ sqs.Queue) {
	_, respErr := sqsQ.SendMessageBatch(msgList)
	if respErr != nil {
		log.Println("ERROR:", respErr)
	}
}
