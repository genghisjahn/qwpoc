package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/sqs"
	"github.com/garyburd/redigo/redis"
)

//Create a redis testJob result with 1,000,000 users.
//Create a redis set for each user, and give each user 0-3 devices
//Devices should have a fake GUID token and a fake GUID endpoint
//Create a group/user user/group with 10,100,1000,10000,1000000 users.
//Create a redis set of some kind that holds all of the message hashes.
//Time how long it takes to do things.

//QuestionQName : name of the problems SQS Queue
const ProblemQName = "demo-problems"

//AnswerQName : name of the answers SQS Queue
const AnswerQName = "demo-answers"

//SQS instance of the sqs var from goamz
var SQS *sqs.SQS

//Question : struct that holds the data for the math problem.
type Question struct {
	Num1 int
	Num2 int
}

var bufferCount = 50 //Playing with this get's different speep through puts.
//The mbAir needs it around 50 or it starts erroring out.
//The mbPro can handle higher values, but diminishing returns.

//But, even set at 50, the mbAir can put 100,000 items in the queu in  48 seconds.

var sem = make(chan bool, bufferCount)
var (
	redisAddress   = flag.String("redis-address", ":6379", "Address to the Redis server")
	maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
)

func main() {
}

func getRedisConn() redis.Conn {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			return nil, err
		}

		return c, err
	}, *maxConnections)
	return redisPool.Get()
}

func createUsers() {
	r := getRedisConn()
	for i := 0; i < 1000000; i++ {
		r.Do("SADD", "all-users-demo", fmt.Sprintf("u:%v", i+1))
	}
}

func getUsers() []string {
	r := getRedisConn()
	userkeys, _ := redis.Strings(r.Do("SMEMBERS", "all-users-demo"))
	return userkeys
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
	probq, getErr := SQS.GetQueue(ProblemQName)
	if getErr != nil {
		fmt.Println(ProblemQName)
		return getErr
	}
	msgSlice := make([]sqs.Message, 0, 10) //A slice that holds up to 10 messages
	msgAll := [][]sqs.Message{}

	for i := 0; i < 100000; i++ {
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
		sem <- true
		go func(sl10 []sqs.Message) {
			addToProblemQ(sl10, *probq)
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
