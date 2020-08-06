package examples

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)
import "github.com/google/uuid"

type Voter struct {
	client *redis.Client
	ctx    *context.Context
}

const (
	Time     int64  = 3600
	Score    int    = 512
	VotedKey string = "voted:"
	ScoreKey string = "score:"
	TimeKey  string = "time"
	VotesKey string = "votes"
	IdKey    string = "id:"
	UserKey  string = "user"
	LinkKey  string = "link"
)

func NewVoter(client *redis.Client) *Voter {
	ctx := context.Background()
	return &Voter{client, &ctx}
}

func (v *Voter) RunVoter() {
	user := "Alex"
	link := "http://github.com/abondar24"
	userId := uuid.New().String()

	v.post(&user, &userId, &link)
	v.vote(&user, &userId)
	v.readPosts()
}

func (v *Voter) vote(user, id *string) {
	cutOff := time.Now().Unix() - Time
	if v.client.ZScore(*v.ctx, TimeKey, *id).Val() < float64(cutOff) {
		return
	}

	if v.client.SAdd(*v.ctx, VotedKey+*id, user).Val() != 0 {
		v.client.ZIncrBy(*v.ctx, ScoreKey, float64(Score), *id)
		v.client.HIncrBy(*v.ctx, VotedKey, *id, 1)
	}

}

func (v *Voter) post(user, id, link *string) {
	voted := VotedKey + *id
	v.client.SAdd(*v.ctx, voted, user)

	//set expire for 3600ms
	v.client.Expire(*v.ctx, voted, time.Second*time.Duration(Time))

	now := time.Now().Unix()
	postId := IdKey + *id

	//create hash
	v.client.HMSet(*v.ctx, postId, map[string]interface{}{
		TimeKey:  now,
		VotesKey: 1,
		UserKey:  *user,
		LinkKey:  *link,
	})

	//post score and time
	v.client.ZAdd(*v.ctx, ScoreKey, &redis.Z{Score: float64(Score + 100), Member: postId})
}

func (v *Voter) readPosts() {

	ids := v.client.ZRevRange(*v.ctx, ScoreKey, 0, 100).Val()
	for _, id := range ids {
		fmt.Println(id)
		fmt.Println(v.client.HGetAll(*v.ctx, id).Val())
	}
}
