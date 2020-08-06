package examples

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)
import "github.com/google/uuid"

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

func (cl *Client) RunVoter() {
	user := "Alex"
	link := "http://github.com/abondar24"
	userId := uuid.New().String()

	cl.post(&user, &userId, &link)
	cl.vote(&user, &userId)
	cl.readPosts()
}

func (cl *Client) vote(user, id *string) {
	cutOff := time.Now().Unix() - Time
	if cl.client.ZScore(*cl.ctx, TimeKey, *id).Val() < float64(cutOff) {
		return
	}

	if cl.client.SAdd(*cl.ctx, VotedKey+*id, user).Val() != 0 {
		cl.client.ZIncrBy(*cl.ctx, ScoreKey, float64(Score), *id)
		cl.client.HIncrBy(*cl.ctx, VotedKey, *id, 1)
	}

}

func (cl *Client) post(user, id, link *string) {
	voted := VotedKey + *id
	cl.client.SAdd(*cl.ctx, voted, user)

	//set expire for 3600ms
	cl.client.Expire(*cl.ctx, voted, time.Second*time.Duration(Time))

	now := time.Now().Unix()
	postId := IdKey + *id

	//create hash
	cl.client.HMSet(*cl.ctx, postId, map[string]interface{}{
		TimeKey:  now,
		VotesKey: 1,
		UserKey:  *user,
		LinkKey:  *link,
	})

	//post score and time
	cl.client.ZAdd(*cl.ctx, ScoreKey, &redis.Z{Score: float64(Score + 100), Member: postId})
}

func (cl *Client) readPosts() {

	ids := cl.client.ZRevRange(*cl.ctx, ScoreKey, 0, 100).Val()
	for _, id := range ids {
		fmt.Println(id)
		fmt.Println(cl.client.HGetAll(*cl.ctx, id).Val())
	}
}
