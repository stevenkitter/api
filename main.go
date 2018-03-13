package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/autotls"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
				"message": "pong",
			})
	})
	log.Fatal(autotls.Run(r, "api.julu666.com", "admin.julu666.com"))
}