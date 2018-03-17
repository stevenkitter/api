package main

import (
	"api/code"
	"api/register"
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

func main() {
	gin.DisableConsoleColor()
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()

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

	//register 注册
	r.POST("/register", register.Register)

	r.Run(":9009")
}
