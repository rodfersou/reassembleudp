package workers

import (
	"context"
	// "fmt"
	"net"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/models"
)

const size = 5000

func ReadUDPWorker(
	id int,
	conn net.PacketConn,
	coll *mongo.Collection,
	ctx context.Context,
	receivingMessage *sync.Map,
) {
	fragments := make(chan *models.Fragment, size)
	go bulkInsertFragment(id, coll, ctx, fragments)
	buf := make([]byte, size)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		fragment := models.CreateFragment(buf)
		(*receivingMessage).Store(fragment.MessageId, time.Now().Unix())
		fragments <- fragment
	}
}

func bulkInsertFragment(
	id int,
	coll *mongo.Collection,
	ctx context.Context,
	fragments <-chan *models.Fragment,
) {
	models := make([]mongo.WriteModel, size)
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
				_, err := coll.BulkWrite(ctx, models[:i], options.BulkWrite().SetOrdered(false))
				if err != nil {
					panic(err)
				}
				i = 0
				models = make([]mongo.WriteModel, size)
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
		if i == size {
			full <- true
			<-done
		}
	}
}
