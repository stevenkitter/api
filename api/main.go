package main

import (
	"api/code"
	"api/common"
	// "api/log"
	"api/login"
	"api/register"
	"api/wx"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"os"
)

func main() {
	gin.DisableConsoleColor()
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	log.SetOutput(gin.DefaultWriter)
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(Secure())
	// log.SetupLog()
	/*
	** 短信验证码接口
	** params: phone
	 */
	r.GET("/code", code.GetCode)
	// 登陆接口
	r.POST("/login", login.Login)
	//register 注册
	r.POST("/register", register.Register)

	//暴露给微信的接口
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "部署成功！",
		})
	})
	//暴露给微信的接口
	r.POST("/wx", wx.WxHandler)

	r.Run(":9009")
}

//CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		headers := "Content-Type, Content-Length, Accept-Encoding," +
			" X-CSRF-Token, Authorization, Nonce, Timestamp, Signature, Token"
		c.Writer.Header().Set("Access-Control-Allow-Headers", headers)
		if c.Request.Method == "OPTIONS" {
			fmt.Println("options")
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

//secure
func Secure() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header
		nonce := header.Get("nonce")
		timestamp := header.Get("timestamp")
		signature := header.Get("signature")
		ret, msg := wx.CheckSignature(timestamp, nonce, signature)
		if !ret {
			common.Fail(c, msg) //
			c.Abort()
			return
		}
		//请求是安全的 来自客户端 并不是绝对安全哦
		c.Next()
	}
}
