package main

import (
	"flag"
	"github.com/abondar24/RedisDemo/examples"
	"log"
)

func main() {
	demoName := flag.String("demo", "", "which example to run")

	pwd := ""
	db := 0
	client := examples.NewClient(&pwd, &db)

	flag.Parse()
	if *demoName == "" {
		log.Fatal("No example provided")
	}

	switch *demoName {
	case "voter":
		client.RunVoter()

	case "token":
		client.RunToken()

	case "tr":
		client.RunTransaction()

	case "log":
		client.RunLogs()

	case "counter":
		client.RunCounter()

	case "ac":
		client.RunAutocomplete()

	case "async":
		client.RunAsync()

	case "queue":
		client.RunQueue()
	}

}
