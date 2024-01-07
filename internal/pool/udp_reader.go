package pool

import (
	"context"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/models"
)

const batch_size = 1000
const queue_size = 10000
const buffer_size = 512

func ReadUDPWorker(
	id int,
	ctx context.Context,
	coll_messages *mongo.Collection,
	coll_fragments *mongo.Collection,
	conn net.PacketConn,
) {
	fragments := make(chan *models.Fragment, queue_size)
	go bulkInsertFragment(id, ctx, coll_messages, coll_fragments, fragments)
	buf := make([]byte, buffer_size)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		fragment := models.CreateFragment(buf)
		fragments <- fragment
	}
}

func bulkInsertFragment(
	id int,
	ctx context.Context,
	coll_messages *mongo.Collection,
	coll_fragments *mongo.Collection,
	fragments <-chan *models.Fragment,
) {
	message_updated_at := make(map[int]bool)
	fragment_batch := make([]mongo.WriteModel, queue_size)
	full := make(chan bool)
	done := make(chan bool)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	i := 0

	go func() {
		is_full := false
		for {
			select {
			case <-full:
				is_full = true
			case <-ticker.C:
				is_full = false
			}
			if i > 0 {
				// Unordered Bulk inserts skip duplicates when the unique index raise error
				_, err := coll_fragments.BulkWrite(ctx, fragment_batch[:i], options.BulkWrite().SetOrdered(false))
				if err != nil {
					panic(err)
				}
				i = 0

				message_batch := make([]mongo.WriteModel, len(message_updated_at))
				index := 0
				for messageId, _ := range message_updated_at {
					delete(message_updated_at, messageId)
					message_batch[index] = mongo.NewUpdateOneModel().SetFilter(
						bson.M{
							"_id": messageId,
						},
					).SetUpdate(
						bson.M{"$setOnInsert": bson.M{
							"_id":        messageId,
							"status":     models.InProgress,
							"updated_at": time.Now(),
						}},
					).SetUpsert(
						true,
					)
					index++
				}
				_, err = coll_messages.BulkWrite(ctx, message_batch, options.BulkWrite().SetOrdered(false))
				if err != nil {
					panic(err)
				}

				if is_full {
					done <- true
				}
			}
		}
	}()

	for fragment := range fragments {
		message_updated_at[fragment.MessageId] = true
		fragment_batch[i] = mongo.NewInsertOneModel().SetDocument(
			fragment,
		)
		i++
		if i == batch_size {
			full <- true
			<-done
		}
	}
}
