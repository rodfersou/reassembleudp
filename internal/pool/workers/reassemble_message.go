package workers

import (
	"context"
	"fmt"
	"golang.org/x/exp/maps"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rodfersou/reassembleudp/internal/models"
	"github.com/rodfersou/reassembleudp/internal/utils"
)

func ReassembleMessageWorker(coll *mongo.Collection, ctx context.Context) {
	mapRetry := make(map[int]bool)
	for {
		retryQueue := maps.Keys(mapRetry)
		sort.Ints(retryQueue[:])
		if len(retryQueue) > 100 {
			delete(mapRetry, retryQueue[0])
			retryQueue = maps.Keys(mapRetry)
			sort.Ints(retryQueue[:])
		}
		var first_fragment models.Fragment
		err := coll.FindOne(
			ctx,
			bson.M{
				"flags":      0,
				"message_id": bson.M{"$nin": retryQueue},
			},
			options.FindOne().SetSort(
				bson.D{
					{"message_id", 1},
					{"offset", 1},
				},
			),
		).Decode(&first_fragment)
		if err != nil {
			// Still not ready
			continue
		}

		filter := bson.M{
			"flags":      0,
			"message_id": first_fragment.MessageId,
			// "created_at": bson.M{"$gte": time.Now().Add(-30 * time.Second)},
		}
		cursor, err := coll.Find(
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
		if len(fragments) == 0 {
			continue
		}

		flags := 0
		holes := utils.ValidateMessage(fragments)
		if len(holes) == 0 {
			flags = 1
			message := utils.ReassembleMessage(fragments)
			hash := utils.HashMessage(message)
			fmt.Printf(
				"Message #%d length: %d sha256:%s\n",
				first_fragment.MessageId,
				len(message),
				hash,
			)
		} else {
			mapRetry[first_fragment.MessageId] = true
			if fragments[0].CreatedAt.Unix() < time.Now().Add(-30*time.Second).Unix() {
				flags = 2
				fmt.Printf(
					"Message #%d Hole at: %d\n",
					first_fragment.MessageId,
					holes[0],
				)
				// for _, hole := range holes {
				//     fmt.Printf(
				//         "Message #%d Hole at: %d\n",
				//         first_fragment.MessageId,
				//         hole,
				//     )
				// }
			}
		}
		_, err = coll.UpdateMany(
			ctx,
			filter,
			bson.M{
				"$set": bson.M{
					"flags": flags,
				},
			},
		)
		if err != nil {
			panic(err)
		}
	}
}
