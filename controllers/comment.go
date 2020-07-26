package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jameslahm/conduit-server-gin/middlewares"
	"github.com/jameslahm/conduit-server-gin/models"
	"github.com/jameslahm/conduit-server-gin/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddCommentInput add comment data
type AddCommentInput struct {
	Body string `json:"string"`
}

// AddComment add comment
func AddComment(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	articleCollection := client.Database("conduit").Collection("articles")
	commentCollection := client.Database("conduit").Collection("comments")

	var loginUser models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	var data AddCommentInput
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var article models.Article
	err = articleCollection.FindOne(ctx, bson.M{
		"slug": c.Param("slug"),
	}).Decode(&article)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	var comment models.Comment
	comment.Body = data.Body
	comment.Article = article.ID
	comment.Author = loginUser.ID
	insertResult, err := commentCollection.InsertOne(ctx, comment)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	comment.ID = insertResult.InsertedID.(primitive.ObjectID)

	var commentJSON models.CommentJSON
	commentJSON.CommentBase = comment.CommentBase

	c.JSON(http.StatusOK, gin.H{
		"comment": commentJSON,
	})
}

// DeleteComment delete comment
func DeleteComment(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	commentCollection := client.Database("conduit").Collection("comments")

	var loginUser models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	_, err = commentCollection.DeleteOne(ctx, bson.M{
		"_id":    c.Param("id"),
		"author": id,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// GetComments get comments
func GetComments(c *gin.Context) {
	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")

	var loginUser models.User
	claims, err := middlewares.Authenticate(c)
	if err == nil {
		id, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&loginUser)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	var article models.Article
	articleCollection := client.Database("conduit").Collection("articles")
	err = articleCollection.FindOne(ctx, bson.M{
		"slug": c.Param("slug"),
	}).Decode(&article)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	commentCollection := client.Database("conduit").Collection("comments")
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "article", Value: article.ID}}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{{
		Key:   "from",
		Value: "users",
	}, {Key: "localField", Value: "author"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "author"}}}}
	cursor, err := commentCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	var comments []models.CommentWithAuthor
	err = cursor.All(ctx, comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	var commentsJSON = make([]models.CommentJSON, len(comments))
	for i := range comments {
		commentsJSON[i].CommentBase = comments[i].CommentBase
		commentsJSON[i].Author = comments[i].Author.ToProfile(&loginUser)
	}
	c.JSON(http.StatusOK, gin.H{
		"comments": commentsJSON,
	})

}
