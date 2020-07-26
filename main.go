package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jameslahm/conduit-server-gin/controllers"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Load .env error")
	}
	r := gin.Default()
	api := r.Group("/api")

	api.POST("/users/login", controllers.Login)
	api.POST("/users", controllers.Register)
	api.GET("/user", controllers.GetCurrentUser)
	api.PUT("/user", controllers.UpdateUser)

	api.GET("/profiles/:username", controllers.GetProfile)
	api.POST("/profiles/:username/follow", controllers.FollowUser)
	api.DELETE("/profiles/:username/follow", controllers.UnFollowUser)

	api.GET("/articles", controllers.GetAllArticles)
	api.GET("/articles/feed", controllers.GetFeedArticles)
	api.GET("/articles/:slug", controllers.GetArticle)
	api.POST("/articles", controllers.CreateArticle)
	api.PUT("/articles/:slug", controllers.UpdateArticle)
	api.DELETE("/articles/:slug", controllers.DeleteArticle)

	api.POST("/articles/:slug/comments", controllers.AddComment)
	api.GET("/articles/:slug/comments", controllers.GetComments)
	api.DELETE("/articles/:slug/comments/:id", controllers.DeleteComment)

	api.POST("/articles/:slug/favorite", controllers.FavoriteArticle)
	api.DELETE("/articles/:slug/favorite", controllers.UnFavoriteArticle)

	api.GET("/tags", controllers.GetTags)

	r.Run()
}
