package v1

import (
	"github.com/stevenkitter/api/v1/common"

	"github.com/gin-gonic/gin"
	// "net/http"
	"strconv"
	"time"
)

//RegisterModel phone code password(md5)
type RegisterModel struct {
	UserPhone string `form:"phone" json:"phone" binding:"required"` //账号 可以是手机号 用户名 邮箱
	Code      string `form:"code" json:"code" binding:"required"`
	Password  string `form:"password" json:"password" binding:"required"` //密码
}
//Register is
func Register(c *gin.Context) {
	// db, dbErr := sql.Open("mysql", "root:julu666@mariadb/julu")
	db, dbErr := common.Mariadb()
	defer db.Close()

	//无法连接数据库
	if dbErr != nil {
		common.Fail(c, dbErr.Error())
	}

	var json RegisterModel
	if jsonErr := c.ShouldBindJSON(&json); jsonErr == nil {
		//判断手机号注册过没
		selectPhone := "SELECT jl_userId FROM JL_User WHERE jl_phone=?"
		var userid string
		db.QueryRow(selectPhone, json.UserPhone).Scan(&userid)

		//用户不存在
		if userid == "" { //userid 值没变 说明 数据库找不到
			//验证码正确
			redis, err := common.Redis()
			defer redis.Close()

			if err != nil {
				common.Fail(c, err.Error())
				return
			}
			reName := json.UserPhone + "code"
			reCode, err := common.RedisString(redis, reName)

			if err != nil && err.Error() == "redigo: nil returned" {
				common.Fail(c, "请发送验证码")
				return
			}
			if err != nil {
				common.Fail(c, err.Error())
				return
			}

			if json.Code != reCode {
				common.Fail(c, "验证码错误")
				return
			} 
			insertSQL := "INSERT INTO JL_User (jl_phone, jl_password, " +
				"jl_register_time, jl_userID) VALUES (?, ?, ?, ?)"
			t := time.Now()
			timestamp := t.Unix()

			userIDStr := strconv.FormatInt(timestamp, 10) + json.UserPhone + common.JL_APPKEY
			userID := common.SecrectKey(userIDStr)
			_, insertErr := db.Exec(insertSQL, json.UserPhone, json.Password, timestamp, userID)
			if insertErr != nil {
				common.Fail(c, insertErr.Error())
				return
			}

			common.OK(c, "注册成功")	
			

		} else {
			common.Fail(c, "用户手机号已注册过")
			return
		}
	} else {
		//json 失效 数据格式不对
		common.Fail(c, jsonErr.Error())
		return
	}
}
