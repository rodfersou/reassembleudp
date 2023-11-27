package pool

import (
	"context"
	"net"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/rodfersou/reassembleudp/internal/pool/workers"
)

const size = 4

func CreatePool(conn net.PacketConn, coll *mongo.Collection, ctx context.Context) {
	// receivingMessage := sync.Map{}
	var wg sync.WaitGroup
	for i := 1; i <= size; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			workers.ReadUDPWorker(i, conn, coll, ctx)
			// workers.ReadUDPWorker(i, conn, coll, ctx, &receivingMessage)
		}()
	}
	// wg.Add(1)
	// go func() {
	//  defer wg.Done()
	//  workers.ReassembleMessageWorker(coll, ctx)
	//  workers.ReassembleMessageWorker(coll, ctx, &receivingMessage)
	// }()
	wg.Wait()
}
