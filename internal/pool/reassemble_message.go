package pool

import (
	"context"
	// "fmt"
	// "time"

	amqp "github.com/rabbitmq/amqp091-go"
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/models"
	// "github.com/rodfersou/reassembleudp/internal/utils"
)

func ReassembleMessageWorker(
	ctx context.Context,
	coll_messages *mongo.Collection,
	coll_fragments *mongo.Collection,
	amqp_ctx context.Context,
	q amqp.Queue,
	ch *amqp.Channel,
) {
	messages := make(map[int]*models.Message)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}
	for d := range msgs {
		fragment := models.CreateFragment(d.Body)
		message, ok := messages[fragment.MessageId]
		if !ok {
			messages[fragment.MessageId] = models.CreateMessage(fragment)
			message = messages[fragment.MessageId]
		}
		message.AddFragment(fragment)
	}
}
