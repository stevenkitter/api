package main

import (
	"api/code"
	"api/common"
	"api/register"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

func main() {
	gin.DisableConsoleColor()
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()
	r.Use(CORSMiddleware())
	/*
	** 短信验证码接口
	** params: phone
	 */
	r.GET("/code", code.GetCode)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	common.Redis()

	//register 注册
	r.POST("/register", register.Register)

	r.Run(":9009")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			fmt.Println("options")
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}
