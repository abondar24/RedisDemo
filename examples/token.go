package examples

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"time"
)

const (
	LoginKey  = "login:"
	RecentKey = "recent:"
	ViewedKey = "viewed:"
)

func (cl *Client) RunToken() {

	user := "Alex"

	item1 := "test1"
	token1 := uuid.New().String()

	item2 := "test2"
	token2 := uuid.New().String()

	item3 := "test3"
	token3 := uuid.New().String()

	item4 := "test4"
	token4 := uuid.New().String()

	cl.updateToken(&token1, &user, &item1)
	cl.updateToken(&token2, &user, &item2)
	cl.updateToken(&token3, &user, &item3)
	cl.updateToken(&token4, &user, &item4)

	cl.checkToken(&token3)

	cl.cleanTokens()
	cl.checkToken(&token4)

}

func (cl *Client) checkToken(token *string) {
	fmt.Println(cl.client.HGet(*cl.ctx, LoginKey, *token))
}

func (cl *Client) updateToken(token, user, item *string) {
	timestamp := time.Now().Unix()

	cl.client.HSet(*cl.ctx, LoginKey, token, user)
	cl.client.ZAdd(*cl.ctx, RecentKey, &redis.Z{Score: float64(1), Member: timestamp})

	if *item != "" {
		//record that item is viewed
		cl.client.ZAdd(*cl.ctx, ViewedKey+*token, &redis.Z{Score: float64(2), Member: timestamp})

		//keep only recent 30
		cl.client.ZRemRangeByRank(*cl.ctx, ViewedKey+*token, 0, -31)
		cl.client.ZIncrBy(*cl.ctx, ViewedKey, -1, *item)
	}
}

func (cl *Client) cleanTokens() {
	size := cl.client.ZCard(*cl.ctx, RecentKey).Val()

	tokens := cl.client.ZRange(*cl.ctx, RecentKey, 0, size).Val()

	var sessionKeys []string

	for _, token := range tokens {
		sessionKeys = append(sessionKeys, token)
	}

	cl.client.Del(*cl.ctx, sessionKeys...)
	cl.client.HDel(*cl.ctx, LoginKey, tokens...)
	cl.client.ZRem(*cl.ctx, RecentKey, tokens)

}
