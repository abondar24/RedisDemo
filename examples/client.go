package examples

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Client struct {
	client *redis.Client
	ctx    *context.Context
}

func NewClient(password *string, db *int) *Client {
	client := initClient(password, db)
	ctx := context.Background()
	return &Client{client, &ctx}
}

func initClient(password *string, db *int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: *password,
		DB:       *db,
	})

}
