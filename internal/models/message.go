package models

import (
	"time"
)

type MessageStatus string

const (
	InProgress MessageStatus = "IN_PROGRESS"
	Valid      MessageStatus = "VALID"
	Invalid    MessageStatus = "INVALID"
)

type Message struct {
	MessageId int           `bson:"_id"        json:"message_id"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
	Status    MessageStatus `bson:"status"     json:"status"`
}
