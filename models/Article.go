package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Article article struct
type Article struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title,omitempty"`
	Description string             `json:"description" bson:"description,omitempty"`
	Body        string             `json:"body" bson:"body,omitempty"`
	TagList     []string           `json:"tagList" bson:"tagList,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"createdAt,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updatedAt,omitempty"`
}



