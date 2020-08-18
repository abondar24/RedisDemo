# Redis Demo

Go-Redis demo (Based on Redis In Action book)

1. Voter: create and post votes with id key. (voter)

2. Token: store ,increment and clean token (token)

3. Transaction: use transaction (tr)

4. Log: store recent log (log)

5. Counter: store and print counter (counter)

6. Autocomplete: autocomplete search (ac)

7. Async: release and acquire locks and semaphores (async)

8. Queue: write and read to queue (queue)

9. Messaging: messaging topic (msg)


## Build and run
```yaml
go get

go build

./RediaDemo --demo=<demo_name_from_brackets>

```

PS some demos have no output so check redis db
