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

// LoginInput login post data
type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login login handler
func Login(c *gin.Context) {

	var data LoginInput
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user models.User
	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	err := userCollection.FindOne(ctx, bson.M{
		"email": data.Email,
	}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := models.VerifyPassword(user.Password, data.Password); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}
	if user.Token, err = models.GenerateJwtToken(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// RegisterInput register post data
type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register register handler
func Register(c *gin.Context) {
	var data RegisterInput
	var err error
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	userCollection := client.Database("conduit").Collection("users")
	defer cancel()
	user := models.User{
		Email:    data.Email,
		Password: data.Password,
		Username: data.Username,
	}
	user.Password = models.GenerateHashPassword(user.Password)

	if _, err := userCollection.InsertOne(ctx, user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	if user.Token, err = models.GenerateJwtToken(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// GetCurrentUser get current user
func GetCurrentUser(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}
	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	var user models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	user.Token, err = models.GenerateJwtToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})

}

// UpdateUserInput update user post data
type UpdateUserInput struct {
	Email    string `json:"email" bson:"email,omitempty"`
	Bio      string `json:"bio" bson:"bio,omitempty"`
	Image    string `json:"image" bson:"image:omitempty"`
	Password string `json:"password" bson:"password,omitempty"`
	Username string `json:"username" bson:"username,omitempty"`
}

// UpdateUser update user
func UpdateUser(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var data UpdateUserInput
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if data.Password != "" {
		data.Password = models.GenerateHashPassword(data.Password)
	}
	userCollection.UpdateOne(ctx, bson.M{
		"_id": id,
	}, bson.M{
		"$set": data,
	})
}
