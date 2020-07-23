package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment comment struct
type Comment struct {
	ID        primitive.ObjectID `bson:"_id" json:"-"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
	Body      string             `bson:"body" json:"body"`
	Author    primitive.ObjectID `bson:"author" json:"author"`
}
