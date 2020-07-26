package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArticleBase article base struct
type ArticleBase struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title          string             `json:"title" bson:"title,omitempty"`
	Slug           string             `bson:"slug"`
	Description    string             `json:"description" bson:"description,omitempty"`
	Body           string             `json:"body" bson:"body,omitempty"`
	TagList        []string           `json:"tagList" bson:"tagList,omitempty"`
	CreatedAt      time.Time          `json:"created_at" bson:"createdAt,omitempty"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updatedAt,omitempty"`
	FavoritesCount int                `json:"favoritesCount" bson:"favoritesCount"`
}

// Article article struct
type Article struct {
	ArticleBase `bson:",inline"`
	Author      primitive.ObjectID `bson:"author"`
}

// ArticleWithAuthor article with author
type ArticleWithAuthor struct {
	ArticleBase `bson:",inline"`
	Author      User `bson:"author"`
}

// ArticleJSON article json
type ArticleJSON struct {
	ArticleBase `bson:",inline"`
	Author Profile `json:"author"`
}
