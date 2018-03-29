package main

import (
	"api/code"
	"api/common"
	"api/login"
	"api/modules"
	"api/register"
	"api/wx"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// Mode = gin.ReleaseMode
	Mode = gin.DebugMode
)

func main() {
	gin.SetMode(Mode)
	gin.DisableConsoleColor()
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	log.SetOutput(gin.DefaultWriter)
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(Secure())

	/*
	** 短信验证码接口
	** params: phone
	 */
	r.GET("/code", code.GetCode)
	r.GET("/pre_auth_code", modules.RequestForPreAuthCode)
	r.GET("/userInfo", modules.UserInfo)

	// 登陆接口
	r.POST("/login", login.Login)

	//register 注册
	r.POST("/register", register.Register)

	//暴露给微信的接口
	r.POST("/wx", wx.WxHandler)

	//测试接口是否成功部署
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "部署成功！",
		})
	})

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
		//验证头部签名
		ret, msg := wx.CheckSignature(timestamp, nonce, signature)
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
			var expires_in int64
			mysql.QueryRow(sql, token).Scan(&expires_in)
			nowTimestamp := time.Now().Unix()
			if nowTimestamp > expires_in {
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
