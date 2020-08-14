package examples

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"log"
	"strings"
)

const (
	ValidChars = "`abcdefghijklmnopqrstuvwxyz{"
	MembersKey = "members:"
)

func (cl *Client) RunAutocomplete() {

	user := "Alex"
	contact := "abondar@mail.com"
	contact1 := "Alex Bondar"

	user1 := "Ali"
	contact2 := "ali@mail.com"

	prefix := "al"
	guild := "g1"

	cl.updateContact(&user, &contact)
	cl.updateContact(&user, &contact1)
	cl.updateContact(&user1, &contact2)

	cl.fetchAutoCompleteList(&user, &prefix)
	cl.joinGuild(&guild, &user)
	cl.autoCompletePrefix(&guild, &prefix)

	cl.leaveGuild(&guild, &guild)
	cl.removeContact(&user, &contact)
	cl.removeContact(&user1, &contact2)
}

func (cl *Client) updateContact(user, contact *string) {
	acList := RecentKey + *user

	pipe := cl.client.Pipeline()

	//remove if exists
	pipe.LRem(*cl.ctx, acList, 1, *contact)

	pipe.RPush(*cl.ctx, acList, *contact)
	pipe.LTrim(*cl.ctx, acList, 0, 99)

	_, err := pipe.Exec(*cl.ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func (cl *Client) removeContact(user, contact *string) {
	cl.client.LRem(*cl.ctx, RecentKey+*user, 1, *contact)
}

func (cl *Client) fetchAutoCompleteList(user, prefix *string) {
	candidates := cl.client.LRange(*cl.ctx, RecentKey+*user, 0, -1).Val()

	for _, cand := range candidates {
		if strings.HasPrefix(strings.ToLower(cand), strings.ToLower(*prefix)) {
			fmt.Println(cand)
		}

	}
}

func (cl *Client) autoCompletePrefix(guild, prefix *string) {

	start, end := cl.findPrefixRange(*prefix)
	id := uuid.New().String()
	start += id
	end += id
	zSetName := MembersKey + *guild
	var items []string

	cl.client.ZAdd(*cl.ctx, zSetName, &redis.Z{Member: start, Score: 0}, &redis.Z{Member: start, Score: 0})

	for {
		err := cl.client.Watch(*cl.ctx, func(tx *redis.Tx) error {
			pipe := tx.Pipeline()
			startIndex := tx.ZRank(*cl.ctx, zSetName, start).Val()
			endIndex := tx.ZRank(*cl.ctx, zSetName, end).Val()
			rng := cl.min(startIndex+9, endIndex-2)

			pipe.ZRem(*cl.ctx, zSetName, start, rng)
			tmp := pipe.ZRange(*cl.ctx, zSetName, startIndex, rng)

			_, err := pipe.Exec(*cl.ctx)
			if err != nil {
				return err
			}

			res := tmp.Val()
			if len(res) != 0 {
				items = res
			}

			return nil

		}, zSetName)
		if err != nil {
			log.Fatal(err)
		}
		break
	}

	for _, item := range items {
		if !strings.Contains(item, "{") {
			fmt.Println(items)
		}

	}
}

func (cl *Client) findPrefixRange(prefix string) (string, string) {
	posn := strings.IndexByte(ValidChars, prefix[len(prefix)-1])
	if posn == 0 {
		posn = 1
	}

	suffix := string(ValidChars[posn-1])

	return prefix[:len(prefix)-1] + suffix + "{", prefix + "{"
}

func (cl *Client) joinGuild(guild, user *string) {
	cl.client.ZAdd(*cl.ctx, MembersKey+*guild, &redis.Z{Member: *user, Score: 0})
}

func (cl *Client) leaveGuild(guild, user *string) {
	cl.client.ZRem(*cl.ctx, MembersKey+*guild, *user)
}

func (cl *Client) min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
