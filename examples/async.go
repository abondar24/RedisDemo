package examples

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"log"
	"math"
	"strconv"
	"time"
)

func (cl *Client) RunAsync() {

	lockname := "lock1"
	acqTime := 100.00
	lockTimeOut := 500.600

	lockId := cl.acquireLock(&lockname, &acqTime, &lockTimeOut)
	fmt.Printf("Acquired lock %s\n", lockId)

	semName := "sem1"
	semTimeOut := int64(100)
	semLimit := int64(10)

	semId := cl.acquireSemaphore(&semName, &semTimeOut, &semLimit)
	fmt.Printf("Acquired semafore %s\n", semId)
	cl.releaseSemaphore(&semName, &semId)

	semId = cl.acquireFairSemaphore(&semName, &semTimeOut, &semLimit)
	fmt.Printf("Acquired fair semafore %s\n", semId)
	cl.refreshFairSemaphore(&semName, &semId)
	cl.releaseFairSemaphore(&semName, &semId)

}

func (cl *Client) acquireLock(lockname *string, acqTimeout, lockTimeout *float64) string {
	id := uuid.New().String()
	*lockTimeout = math.Ceil(*lockTimeout)
	end := time.Now().UnixNano() + int64(*acqTimeout)

	for time.Now().UnixNano() < end {
		if cl.client.SetNX(*cl.ctx, *lockname, id, 0).Val() {
			cl.client.Expire(*cl.ctx, *lockname, time.Duration(*lockTimeout)*time.Second)
			return id
		} else if cl.client.TTL(*cl.ctx, *lockname).Val() < 0 {
			cl.client.Expire(*cl.ctx, *lockname, time.Duration(*lockTimeout)*time.Second)
		}

		time.Sleep(10 * time.Millisecond)
	}

	return "Lock already acquired"
}

func (cl *Client) acquireSemaphore(semaphoreName *string, limit *int64, timeout *int64) string {
	id := uuid.New().String()
	now := time.Now().UnixNano()

	pipe := cl.client.TxPipeline()
	pipe.ZRemRangeByScore(*cl.ctx, *semaphoreName, "-inf", strconv.Itoa(int(now-*timeout)))
	pipe.ZAdd(*cl.ctx, *semaphoreName, &redis.Z{Member: id, Score: float64(now)})
	res := pipe.ZRank(*cl.ctx, *semaphoreName, id)

	_, err := pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatal(err)
	}

	if res.Val() < *limit {
		return id
	}

	cl.client.ZRem(*cl.ctx, *semaphoreName, id)
	return "Semaphore already acquired"
}

//fair is not dependent on system time
func (cl *Client) acquireFairSemaphore(semaphoreName *string, limit *int64, timeout *int64) string {
	id := uuid.New().String()
	now := time.Now().UnixNano()

	czset := *semaphoreName + ":owner"
	counter := *semaphoreName + ":counter"

	pipe := cl.client.TxPipeline()
	pipe.ZRemRangeByScore(*cl.ctx, *semaphoreName, "-inf", strconv.Itoa(int(now-*timeout)))
	pipe.ZInterStore(*cl.ctx, czset, &redis.ZStore{Keys: []string{czset, *semaphoreName}, Weights: []float64{1, 0}})

	_, err := pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatal(err)
	}

	ctr := pipe.Incr(*cl.ctx, counter).Val()

	pipe.ZAdd(*cl.ctx, *semaphoreName, &redis.Z{Member: id, Score: float64(now)})
	pipe.ZAdd(*cl.ctx, czset, &redis.Z{Member: id, Score: float64(ctr)})

	res := pipe.ZRank(*cl.ctx, *semaphoreName, id)

	if res.Val() < *limit {
		return id
	}

	cl.client.ZRem(*cl.ctx, *semaphoreName, id)
	cl.client.ZRem(*cl.ctx, czset, id)

	_, err = pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatal(err)
	}

	return "Semaphore already acquired"
}

func (cl *Client) releaseSemaphore(semaphoreName, id *string) {
	cl.client.ZRem(*cl.ctx, *semaphoreName, *id)
}

func (cl *Client) releaseFairSemaphore(semaphoreName, id *string) {
	pipe := cl.client.TxPipeline()
	pipe.ZRem(*cl.ctx, *semaphoreName, *id)
	pipe.ZRem(*cl.ctx, *semaphoreName+":owner", *id)

	_, err := pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatalln(err)
	}
}

func (cl *Client) refreshFairSemaphore(semaphoreName, id *string) {
	res := cl.client.ZAdd(*cl.ctx, *semaphoreName, &redis.Z{Member: *id, Score: float64(time.Now().Unix())}).Val()

	if res != 0 {
		cl.releaseSemaphore(semaphoreName, id)
		fmt.Printf("Failed to refresh semaphore : %s\n", *id)
	} else {
		fmt.Printf("Semafore refreshed: %s\n", *id)
	}
}
