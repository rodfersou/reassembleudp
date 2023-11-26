package pool

import (
	"context"
	"net"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/rodfersou/reassembleudp/internal/pool/workers"
)

func CreatePool(conn net.PacketConn, coll *mongo.Collection, ctx context.Context) {
	var wg sync.WaitGroup
	for i := 1; i <= 4; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			workers.ReadUDPWorker(i, conn, coll, ctx)
		}()
	}
	// go func() {
	//  defer wg.Done()
	//  workers.ReassembleMessageWorker(coll, ctx)
	// }()
	wg.Wait()
}
