// 注册
package register

import (
	"github.com/gin-gonic/gin"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

type RegisterModel struct {
	UserName     string `form:"username" json:"username" binding:"required"` //账号 可以是手机号 用户名 邮箱
	Code	string `form:"code" json:"code" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"` //密码 
}

func Register(c *gin.Context) {
	var json RegisterModel
	db, err := sql.Open("mysql", "root:julu666@/julu")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	if err := c.ShouldBindJSON(&json); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
	}
}

