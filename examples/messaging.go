package examples

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"time"
)

const (
	IdsKey     = "ids:"
	TopicKey   = "topic:"
	SeenKey    = "seen:"
	MessageKey = "msgs:"
)

type Message struct {
	Id        int64
	Timestamp int64
	Sender    string
	Message   string
}

func (cl *Client) RunMessaging() {
	sender := "sender1"

	var recipients []string
	rec1 := "r1"
	rec2 := "r2"
	recipients = append(recipients, rec1)
	recipients = append(recipients, rec2)

	topic := "test"

	cl.createTopic(&sender, &recipients, &topic)
	cl.joinTopic(&topic, &rec1)
	cl.joinTopic(&topic, &rec2)

	msg := "Hello World!"
	cl.sendMessage(&topic, &sender, &msg)
	cl.fetchPending(&rec1)
	cl.fetchPending(&rec2)

	cl.leaveTopic(&topic, &rec1)

}

func (cl *Client) createTopic(sender *string, recipients *[]string, topic *string) {
	if *topic == "" {
		*topic = strconv.Itoa(int(cl.client.Incr(*cl.ctx, IdsKey+TopicKey).Val()))
	}

	*recipients = append(*recipients, *sender)

	var recipZ []*redis.Z

	for _, r := range *recipients {
		rz := &redis.Z{Member: r, Score: 0}
		recipZ = append(recipZ, rz)
	}

	pipe := cl.client.TxPipeline()
	pipe.ZAdd(*cl.ctx, TopicKey+*topic, recipZ...)

	for _, r := range *recipients {
		pipe.ZAdd(*cl.ctx, SeenKey+r, &redis.Z{Member: *topic, Score: 0})
	}

	_, err := pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatalln(err)
	}

}

func (cl *Client) sendMessage(topicId, sender, msg *string) {
	lockName := TopicKey + *topicId
	timeout := 100.00

	lockId := cl.acquireLock(&lockName, &timeout, &timeout)
	if lockId == "" {
		log.Println("Couldn't acquire lock")
		return
	}

	msgId := cl.client.Incr(*cl.ctx, lockName)
	ts := time.Now().Unix()

	message := &Message{
		Id:        msgId.Val(),
		Timestamp: ts,
		Sender:    *sender,
		Message:   *msg,
	}

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	cl.client.ZAdd(*cl.ctx, MessageKey+*topicId, &redis.Z{Member: jsonMsg, Score: float64(msgId.Val())})
}

func (cl *Client) fetchPending(recipient *string) {
	seen := cl.client.ZRangeWithScores(*cl.ctx, SeenKey+*recipient, 0, -1).Val()

	pipe := cl.client.TxPipeline()

	res := &redis.StringSliceCmd{}
	tmp := make([]string, 0, len(seen))

	for _, val := range seen {
		topicId := val.Member.(string)
		seenId := val.Score

		res = pipe.ZRange(*cl.ctx, MessageKey+topicId, int64(seenId), 1000)
		tmp = append(tmp, topicId, strconv.Itoa(int(seenId)))
	}

	_, err := pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatalln(err)
	}

	topicInfo := [][]string{tmp, res.Val()}

	for i := 0; i < len(topicInfo); i += 2 {

		message := Message{}
		for _, v := range topicInfo[i+1] {

			if err := json.Unmarshal([]byte(v), &message); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Recieved message: ", message.Message)
		}

		topicId := topicInfo[i][0]
		seenId := float64(message.Id)

		//clean read messages
		cl.client.ZAdd(*cl.ctx, TopicKey+topicId, &redis.Z{Member: recipient, Score: seenId})
		minId := cl.client.ZRangeWithScores(*cl.ctx, TopicKey+topicId, 0, 0).Val()

		pipe.ZAdd(*cl.ctx, SeenKey+*recipient, &redis.Z{Member: topicId, Score: seenId})
		if minId != nil {
			pipe.ZRemRangeByScore(*cl.ctx, MessageKey+topicId, string(rune(0)), strconv.Itoa(int(minId[0].Score)))
		}

	}

}

func (cl *Client) joinTopic(topic, consumer *string) {
	msgId, _ := cl.client.Get(*cl.ctx, IdsKey).Float64()

	pipe := cl.client.TxPipeline()
	pipe.ZAdd(*cl.ctx, TopicKey+*topic, &redis.Z{Member: *consumer, Score: msgId})
	pipe.ZAdd(*cl.ctx, SeenKey+*consumer, &redis.Z{Member: *topic, Score: msgId})

	_, err := pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatalln(err)
	}
}

func (cl *Client) leaveTopic(topic, consumer *string) {
	pipe := cl.client.TxPipeline()
	pipe.ZRem(*cl.ctx, TopicKey+*topic, *consumer)
	pipe.ZRem(*cl.ctx, SeenKey+*consumer, *topic)
	res := pipe.ZCard(*cl.ctx, TopicKey+*topic)

	_, err := pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatalln(err)
	}

	if res == nil {
		pipe.Del(*cl.ctx, MessageKey+*topic)
		pipe.Del(*cl.ctx, MessageKey+*topic)

		_, err := pipe.Exec(*cl.ctx)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		oldest := cl.client.ZRangeWithScores(*cl.ctx, TopicKey+*topic, 0, 0).Val()[0]
		cl.client.ZRemRangeByScore(*cl.ctx, TopicKey+*topic, "0", strconv.Itoa(int(oldest.Score)))
	}

}
