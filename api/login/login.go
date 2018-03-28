// 用户登陆接口
package login

import (
	"api/common"
	"api/wx"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
	// "log"
)

//phone code password(md5)
type LoginModel struct {
	Phone    string `form:"phone" json:"phone" binding:"required"`       //账号 手机号
	Password string `form:"password" json:"password" binding:"required"` //密码
}

type ResultLoginModel struct {
	Phone  string `form:"jl_phone" json:"phone" binding:"required"`
	UserId string `form:"jl_userID" json:"userId" binding:"required"`
	Name   string `form:"jl_name" json:"name" binding:"required"`
}

func Login(c *gin.Context) {
	db, dbErr := common.Mariadb()
	defer db.Close()

	//无法连接数据库
	if dbErr != nil {
		common.Fail(c, dbErr.Error())
	}
	var json LoginModel
	if jsonErr := c.ShouldBindJSON(&json); jsonErr == nil {
		pre, err := db.Prepare(`SELECT jl_password FROM JL_User WHERE jl_phone = ?`)
		defer pre.Close()
		checkErr(c, err)

		// detail, err := db.Prepare(`SELECT jl_phone, jl_userID, jl_name FROM JL_User WHERE jl_phone = ?`)
		// defer detail.Close()
		// checkErr(c, err)

		inToken, err := db.Prepare(`UPDATE JL_User SET token = ?, token_expires_in = ? WHERE jl_phone = ?`)
		defer inToken.Close()
		checkErr(c, err)

		var tokenStr = json.Phone + json.Password + wx.Content_key
		token := common.Sha1(tokenStr)
		h, _ := time.ParseDuration("1h")
		token_expires_in_time := time.Now().Add(h * 2)
		token_expires_in_timestamp := token_expires_in_time.Unix()
		token_expires_in := strconv.FormatInt(token_expires_in_timestamp, 10)
		_, err = inToken.Exec(token, token_expires_in, json.Phone)
		if err != nil {
			common.Fail(c, err.Error())
			return
		}

		var passW string
		pre.QueryRow(json.Phone).Scan(&passW)

		if passW == "" {
			//无用户
			common.Fail(c, "用户不存在，欢迎您注册。")
			return
		} else {
			// log.Printf("passw is %s json.Pass %s", passW, json.Password)
			if passW == json.Password {

				data := map[string]interface{}{"token": token}
				common.OKWithData(c, "登陆成功", data)
			} else {
				common.Fail(c, "密码错误")
				return
			}
		}

	} else {
		common.Fail(c, jsonErr.Error())
		return
	}
}

func checkErr(c *gin.Context, err error) {
	if err != nil {
		common.Fail(c, err.Error())
		panic(err)
	}
}
