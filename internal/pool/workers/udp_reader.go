package workers

import (
	"context"
	// "fmt"
	"net"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/models"
)

func ReadUDPWorker(
	id int,
	conn net.PacketConn,
	coll *mongo.Collection,
	ctx context.Context,
) {
	fragments := make(chan *models.Fragment, 1024)
	go dbInserter(id, coll, ctx, fragments)
	buf := make([]byte, 1024)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		fragments <- models.CreateFragment(buf)
	}
}

func dbInserter(id int, coll *mongo.Collection, ctx context.Context, fragments <-chan *models.Fragment) {
	models := make([]mongo.WriteModel, 1024)
	i := 0
	// tid := 1
	for fragment := range fragments {
		models[i] = mongo.NewInsertOneModel().SetDocument(
			fragment,
		)

		i++
		if i == 1024 {
			// Unordered Bulk inserts skip duplicates when the unique index raise error
			_, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
			if err != nil {
				panic(err)
			}
			// tid = fragment.TransactionId
			i = 0
			models = make([]mongo.WriteModel, 1024)
		}
	}
	// if i > 0 {
	//     // Unordered Bulk inserts skip duplicates when the unique index raise error
	//     _, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	//     if err != nil {
	//         panic(err)
	//     }
	//     fmt.Println("Last transaction ", tid)
	// }
}
