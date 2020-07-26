package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// CommentBase comment base
type CommentBase struct {
	ID        primitive.ObjectID `bson:"_id" json:"-"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
	Body      string             `bson:"body" json:"body"`
	Article   primitive.ObjectID `bson:"article" json:"-"`
}

// Comment comment struct
type Comment struct {
	CommentBase
	Author primitive.ObjectID `bson:"author" json:"author"`
}

// CommentWithAuthor comment with author
type CommentWithAuthor struct {
	CommentBase
	Author User `bson:"author"`
}

// CommentJSON comment with json
type CommentJSON struct {
	CommentBase
	Author Profile `json:"author"`
}
