package examples

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"sort"
	"strconv"
	"time"
)

var Precision = [...]int64{1, 5, 60, 300, 3600, 18000, 86400}

const (
	KnownKey = "known:"
	CountKey = "count:"
)

func (cl *Client) RunCounter() {

	name := "Alex"
	cl.updateCounter(&name)
	cl.updateCounter(&name)
	cl.updateCounter(&name)
	cl.updateCounter(&name)

	precision := 3600
	cl.getCount(&name, &precision)
}

func (cl *Client) updateCounter(name *string) {
	now := time.Now().Local().Unix()
	pipe := cl.client.Pipeline()

	for _, pr := range Precision {
		pNow := now / pr * pr

		hash := fmt.Sprintf("%d:%s", pr, *name)

		pipe.ZAdd(*cl.ctx, KnownKey, &redis.Z{Member: hash, Score: 0})
		pipe.HIncrBy(*cl.ctx, CountKey+hash, strconv.Itoa(int(pNow)), 1)
	}

	_, err := pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func (cl *Client) getCount(name *string, prec *int) {
	hash := fmt.Sprintf("%d:%s", *prec, *name)
	data := cl.client.HGetAll(*cl.ctx, CountKey+hash).Val()
	res := make([][]int, 0, len(data))

	for k, v := range data {
		temp := make([]int, 2)
		key, _ := strconv.Atoi(k)
		val, _ := strconv.Atoi(v)
		temp[0], temp[1] = key, val
		res = append(res, temp)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i][0] < res[j][0]
	})

	fmt.Println(res)
}
