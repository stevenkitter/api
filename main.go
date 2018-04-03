package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stevenkitter/api/v1"
	"github.com/stevenkitter/api/v1/common"
	"github.com/stevenkitter/api/v1/websocket"
)

const (
	//Mode 开发模式
	Mode = gin.ReleaseMode
	// Mode = gin.DebugMode
)

func main() {
	gin.SetMode(Mode)
	gin.DisableConsoleColor()
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	log.SetOutput(gin.DefaultWriter)
	r := gin.Default()
	r.LoadHTMLFiles("./static/index.tmpl")

	r.Use(CORSMiddleware())
	r.Use(Secure())

	v1g := r.Group("/v1")
	v1Setup(v1g)

	hub := websocket.NewHub()
	go hub.Run()
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c.Writer, c.Request)
	})

	r.GET("/socket", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.HTML(http.StatusOK, "index.tmpl", nil)
	})
	r.Run(":8888")
}

/*
** 第一个版本 以后版本大的升级就这么弄
** params: version 1.0
 */
func v1Setup(v1g *gin.RouterGroup) {
	//验证码接口
	v1g.GET("/code", v1.GetCode)

	// 登陆平台接口
	v1g.POST("/login", v1.Login)

	//register 注册用户
	v1g.POST("/register", v1.Register)

	//测试接口是否成功部署
	v1g.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "部署成功！",
		})
	})

}

//CORSMiddleware is middle
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

//Secure is for security
func Secure() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header
		nonce := header.Get("nonce")
		timestamp := header.Get("timestamp")
		signature := header.Get("signature")
		//验证头部签名
		ret, msg := v1.CheckSignature(timestamp, nonce, signature)
		if !ret && Mode == gin.ReleaseMode {
			common.Fail(c, msg) //
			c.Abort()
			return
		}
		//验证token有效性
		token := header.Get("token")
		//为空说明是登陆 无所谓了 不判断token是否失效
		if token != "" {
			mysql, err := common.Mariadb()
			if err != nil {
				common.Fail(c, err.Error())
				c.Abort()
				return
			}
			defer mysql.Close()

			sql := "SELECT token_expires_in FROM JL_User WHERE token = ?"
			var expiresIn int64
			mysql.QueryRow(sql, token).Scan(&expiresIn)
			nowTimestamp := time.Now().Unix()
			if nowTimestamp > expiresIn {
				//过期了
				common.FailToken(c, "token失效")
				c.Abort()
				return
			}
		}

		//请求是安全的 来自客户端 并不是绝对安全哦
		c.Next()
	}
}
