package pool

import (
	"context"
	"net"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

const size = 4

func CreatePool(
	conn net.PacketConn,
	coll_messages *mongo.Collection,
	coll_fragments *mongo.Collection,
	ctx context.Context,
) {
	var wg sync.WaitGroup
	for i := 1; i <= size; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			ReadUDPWorker(i, conn, coll_messages, coll_fragments, ctx)
		}()
	}
	// wg.Add(1)
	// go func() {
	//  defer wg.Done()
	//  ReassembleMessageWorker(coll_messages, coll_fragments, ctx)
	// }()
	wg.Wait()
}
