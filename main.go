package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Load .env error")
	}
	r := gin.Default()
	api := r.Group("/api")

	api.GET("/")

	r.Run()
}
