package workers

import (
	"context"
	// "fmt"
	"net"
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
) {
	fragments := make(chan *models.Fragment, size)
	go bulkInsertFragment(id, coll, ctx, fragments)
	buf := make([]byte, size)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		fragments <- models.CreateFragment(buf)
	}
}

func bulkInsertFragment(id int, coll *mongo.Collection, ctx context.Context, fragments <-chan *models.Fragment) {
	models := make([]mongo.WriteModel, size)
	i := 0

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			<-ticker.C
			if i > 0 {
				// Unordered Bulk inserts skip duplicates when the unique index raise error
				_, err := coll.BulkWrite(ctx, models[:i], options.BulkWrite().SetOrdered(false))
				if err != nil {
					// panic(err)
				}
				// fmt.Println("Ticker ", i)
				i = 0
				models = make([]mongo.WriteModel, size)
			}
		}
	}()

	for fragment := range fragments {
		models[i] = mongo.NewInsertOneModel().SetDocument(
			fragment,
		)

		i++
		if i == size {
			// Unordered Bulk inserts skip duplicates when the unique index raise error
			_, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
			if err != nil {
				// panic(err)
			}
			i = 0
			models = make([]mongo.WriteModel, size)
		}
	}
}
