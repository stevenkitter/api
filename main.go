package main

import (
	"os"
	"github.com/gin-gonic/gin"
	"app/user/register"
)

func main() {
	gin.DisableConsoleColor()
	f, _ := os.Create("gin.log")
    gin.DefaultWriter = io.MultiWriter(f, os.Stdout)


	r := gin.Default()
	//test
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
				"message": "pong",
			})
	})

	//register 注册
	r.POST("/register", register.Register)


	r.Run(":9009")
}