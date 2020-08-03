package main

import (
	"flag"
	"github.com/go-redis/redis/v8"
	"log"
)

func main() {
	demoName := flag.String("demo", "", "which example to run")

	flag.Parse()
	if *demoName == "" {
		log.Fatal("No example provided")
	}
}

func initClient(password *string, db *int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: *password,
		DB:       *db,
	})

}
