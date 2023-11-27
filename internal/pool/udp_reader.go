package pool

import (
	"context"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/models"
)

const batch_size = 4000
const buffer_size = 512

func ReadUDPWorker(
	id int,
	ctx context.Context,
	coll_messages *mongo.Collection,
	coll_fragments *mongo.Collection,
	conn net.PacketConn,
) {
	fragments := make(chan *models.Fragment, batch_size)
	go bulkInsertFragment(id, coll_fragments, ctx, fragments)
	buf := make([]byte, buffer_size)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		fragment := models.CreateFragment(buf)
		// (*receivingMessage).Store(fragment.MessageId, time.Now().Unix())
		fragments <- fragment
	}
}

func bulkInsertFragment(
	id int,
	coll_fragments *mongo.Collection,
	ctx context.Context,
	fragments <-chan *models.Fragment,
) {
	models := make([]mongo.WriteModel, batch_size)
	full := make(chan bool)
	done := make(chan bool)
	ticker := time.NewTicker(5 * time.Second)
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
				_, err := coll_fragments.BulkWrite(ctx, models[:i], options.BulkWrite().SetOrdered(false))
				if err != nil {
					panic(err)
				}
				i = 0
				models = make([]mongo.WriteModel, batch_size)
				if is_full {
					done <- true
				}
			}
		}
	}()

	for fragment := range fragments {
		models[i] = mongo.NewInsertOneModel().SetDocument(
			fragment,
		)
		i++
		if i == batch_size {
			full <- true
			<-done
		}
	}
}
