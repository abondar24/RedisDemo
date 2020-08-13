package examples

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"log"
	"time"
)

const (
	ItemKey = "item:"
)

func (cl *Client) RunTransaction() {

	id := uuid.New().String()
	item := "item1"

	err := cl.trAdd(id, item)
	if err != nil {
		log.Fatal(err)
	}
}

func (cl *Client) trAdd(id, item string) error {
	for time.Now().Unix() < time.Now().Unix()+10 {

		//watch if item assigned to id
		return cl.client.Watch(*cl.ctx, func(tx *redis.Tx) error {
			if _, err := tx.TxPipelined(*cl.ctx, func(pipeliner redis.Pipeliner) error {
				if !tx.SIsMember(*cl.ctx, item, id).Val() {
					//unwatch if not
					tx.Unwatch(*cl.ctx, item)
				}
				pipeliner.ZAdd(*cl.ctx, ItemKey, &redis.Z{Member: item, Score: float64(Score)})
				pipeliner.SRem(*cl.ctx, item, id)

				return nil
			}); err != nil {
				return err
			}
			return nil
		}, item)
	}

	return nil
}
