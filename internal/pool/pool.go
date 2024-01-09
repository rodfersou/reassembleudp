package pool

import (
	"context"
	"net"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
)

const size = 4

func CreatePool(
	ctx context.Context,
	coll_messages *mongo.Collection,
	coll_fragments *mongo.Collection,
	conn net.PacketConn,
	amqp_ctx context.Context,
	q amqp.Queue,
	ch *amqp.Channel,
) {
	var wg sync.WaitGroup
	for i := 1; i <= size; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			ReadUDPWorker(
				i,
				conn,
				amqp_ctx,
				q,
				ch,
			)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ReassembleMessageWorker(
			ctx,
			coll_messages,
			coll_fragments,
			amqp_ctx,
			q,
			ch,
		)
	}()
	wg.Wait()
}
