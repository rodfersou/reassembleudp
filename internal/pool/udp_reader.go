package pool

import (
	"context"
	"net"
	// "time"

	amqp "github.com/rabbitmq/amqp091-go"
	// "github.com/rodfersou/reassembleudp/internal/models"
)

const batch_size = 1000
const queue_size = 10000
const buffer_size = 512

func ReadUDPWorker(
	id int,
	conn net.PacketConn,
	amqp_ctx context.Context,
	q amqp.Queue,
	ch *amqp.Channel,
) {
	buf := make([]byte, buffer_size)
	for {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		err = ch.PublishWithContext(
			amqp_ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        buf,
			},
		)
		if err != nil {
			panic(err)
		}
	}
}
