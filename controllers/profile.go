package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jameslahm/conduit-server-gin/middlewares"
	"github.com/jameslahm/conduit-server-gin/models"
	"github.com/jameslahm/conduit-server-gin/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetProfile get profile
func GetProfile(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	var loginUser models.User
	var user models.User

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")

	if err == nil {
		id, err := primitive.ObjectIDFromHex(claims.UserID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		userCollection.FindOne(ctx, bson.M{
			"_id": id,
		}).Decode(loginUser)
	}

	err = userCollection.FindOne(ctx, bson.M{
		"username": c.Param("username"),
	}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"profile": user.ToProfile(&loginUser),
	})
}

// FollowUser follow user
func FollowUser(c *gin.Context) {
	var loginUser models.User
	var user models.User

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

	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = userCollection.FindOne(ctx, bson.M{
		"username": c.Param("username"),
	}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	loginUser.Follow(&user)
	updateResult, err := userCollection.UpdateOne(ctx, bson.M{
		"_id": loginUser.ID,
	}, bson.M{
		"$set": loginUser,
	})
	if err != nil || updateResult.ModifiedCount != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"profile": user.ToProfile(&loginUser),
	})
}

// UnFollowUser unfollow user
// TODO: Refactor!
func UnFollowUser(c *gin.Context){
	var loginUser models.User
	var user models.User

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

	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = userCollection.FindOne(ctx, bson.M{
		"username": c.Param("username"),
	}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	loginUser.UnFollow(&user)
	updateResult, err := userCollection.UpdateOne(ctx, bson.M{
		"_id": loginUser.ID,
	}, bson.M{
		"$set": loginUser,
	})
	if err != nil || updateResult.ModifiedCount != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"profile": user.ToProfile(&loginUser),
	})
}