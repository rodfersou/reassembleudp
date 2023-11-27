package pool

import (
	"context"
	"net"
	"sync"
	// "time"

	"go.mongodb.org/mongo-driver/mongo"
)

const size = 4

func CreatePool(
	ctx context.Context,
	coll_messages *mongo.Collection,
	coll_fragments *mongo.Collection,
	conn net.PacketConn,
) {
	// ticker := time.NewTicker(3 * time.Second)
	// defer ticker.Stop()
	// go func() {
	//     ReassembleMessageWorker(ctx, coll_messages, coll_fragments)
	//     <-ticker.C
	// }()

	var wg sync.WaitGroup
	for i := 1; i <= size; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			ReadUDPWorker(i, ctx, coll_messages, coll_fragments, conn)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ReassembleMessageWorker(ctx, coll_messages, coll_fragments)
	}()
	wg.Wait()
}
