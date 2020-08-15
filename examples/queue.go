package examples

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"time"
)

const (
	QueueKey   = "queue:"
	QueueName  = QueueKey + "test"
	DelayedKey = "delayed"
)

func (cl *Client) RunQueue() {

	item := "Something"
	cl.addToQueue(&item)
	cl.processQueue()

	queue := "delay"
	delay := int64(100)
	cl.executeLater(&queue, &item, &delay)
}

func (cl *Client) addToQueue(item *string) {
	cl.client.RPush(*cl.ctx, QueueName, *item)
}

func (cl *Client) processQueue() {
	for {
		packed := cl.client.BLPop(*cl.ctx, time.Duration(5)*time.Second, QueueName).Val()

		if len(packed) != 0 {
			fmt.Printf("Retrieved from queue: %s\n", packed[1])

			break
		}
	}
}

func (cl *Client) executeLater(queue, item *string, delay *int64) string {
	id := uuid.New().String()

	if *delay > 0 {
		cl.client.ZAdd(*cl.ctx, DelayedKey+*item, &redis.Z{Member: id, Score: float64(time.Now().Unix() + *delay)})
	} else {
		cl.client.RPush(*cl.ctx, QueueKey+*queue, *item)
	}

	return id
}
