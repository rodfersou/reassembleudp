package pool

import (
	"context"
	"net"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

const size = 4

func CreatePool(
	ctx context.Context,
	coll_messages *mongo.Collection,
	coll_fragments *mongo.Collection,
	conn net.PacketConn,
) {
	var wg sync.WaitGroup
	for i := 1; i <= size; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			ReadUDPWorker(i, ctx, coll_messages, coll_fragments, conn)
		}()
	}
	ReassembleMessageWorker(ctx, coll_messages, coll_fragments)
	wg.Wait()
}
