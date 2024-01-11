package pool

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/models"
	"github.com/rodfersou/reassembleudp/internal/utils"
)

func ReassembleMessageWorker(
	ctx context.Context,
	coll_messages *mongo.Collection,
	coll_fragments *mongo.Collection,
) {
	for {
		cursor, err := coll_messages.Find(
			ctx,
			bson.M{
				"status": models.InProgress,
			},
			options.Find().SetSort(
				bson.D{
					{"updated_at", 1},
					{"_id", 1},
				},
			),
		)
		if err != nil {
			panic(err)
		}
		var messages []models.Message
		if err = cursor.All(ctx, &messages); err != nil {
			panic(err)
		}
		for _, message := range messages {
			cursor.Decode(&message)
			if message.Status != models.InProgress {
				continue
			}
			if message.Status != models.InProgress {
				continue
			}
			filter := bson.M{
				"message_id": message.MessageId,
			}
			cursor, err := coll_fragments.Find(
				ctx,
				filter,
				options.Find().SetSort(
					bson.D{
						{"message_id", 1},
						{"offset", 1},
					},
				),
			)
			if err != nil {
				panic(err)
			}
			var fragments []models.Fragment
			if err = cursor.All(ctx, &fragments); err != nil {
				panic(err)
			}

			holes := utils.ValidateMessage(fragments)
			status := models.InProgress
			if len(holes) == 0 {
				status = models.Valid
				data := utils.ReassembleMessage(fragments)
				checksum := utils.HashMessage(data)
				fmt.Printf(
					"Message #%d length: %d sha256:%s\n",
					message.MessageId,
					len(data),
					checksum,
				)
			} else {
				if message.UpdatedAt.Unix() < time.Now().Add(-30*time.Second).Unix() {
					status = models.Invalid
					fmt.Printf(
						"Message #%d Hole at: %d\n",
						message.MessageId,
						holes[0],
					)
					// for _, hole := range holes {
					//  fmt.Printf(
					//      "Message #%d Hole at: %d\n",
					//      message.MessageId,
					//      hole,
					//  )
					// }
				}
			}
			if status != models.InProgress {
				message.Status = status
				_, err := coll_messages.UpdateOne(
					ctx,
					bson.M{"_id": message.MessageId},
					bson.M{"$set": bson.M{
						"status": status,
					}},
				)
				if err != nil {
					panic(err)
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
