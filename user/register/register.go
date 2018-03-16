// 注册
package register

import (
	"github.com/gin-gonic/gin"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

//phone code password(md5)
type RegisterModel struct {
	UserPhone     string `form:"phone" json:"phone" binding:"required"` //账号 可以是手机号 用户名 邮箱
	Code	string `form:"code" json:"code" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"` //密码 
}

func Register(c *gin.Context) {
	// db, dbErr := sql.Open("mysql", "root:julu666@mariadb/julu")
	db, dbErr := sql.Open("mysql", "root:julu666@tcp(115.159.222.199:3306)/julu")
	defer db.Close()

	//无法连接数据库
	if dbErr != nil {
		fail(c, dbErr.Error())
	}

	var json RegisterModel
	if jsonErr := c.ShouldBindJSON(&json); jsonErr == nil {
		//判断手机号注册过没
		select_phone := "SELECT jl_userId FROM JL_User WHERE jl_phone=?"
		var userid string
		db.QueryRow(select_phone, json.UserPhone).Scan(&userid)
		
		//用户不存在
		if userid == "" {
			//验证码正确
			if json.Code == "123" {
				insertSql := "INSERT INTO JL_User (jl_phone, jl_password) VALUES (?, ?)"
				re, insertErr := db.Exec(insertSql, json.UserPhone, json.Password)
				if insertErr != nil {
					fail(c, insertErr.Error())
					return
				}

				id, lastErr := re.LastInsertId()

				if lastErr != nil {
					fail(c, lastErr.Error())
					return
				}
				c.JSON(http.StatusOK, gin.H{"userid": id})
				return
			}else{
				fail(c, "验证码错误")
				return
			}
			
		}else{
			fail(c, "用户手机号已注册过")
			return
		}
	}else{
		//json 失效 数据格式不对
		fail(c, jsonErr.Error())
		return
	}
}

func fail(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, gin.H{"jl_error": err})
	return
}

