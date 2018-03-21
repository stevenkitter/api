package common

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
)

const (
	JL_APPKEY = "58cde7973b53b19c114f39cd9936c25c37943c6a"                    //appkey
	RES_OK    = 0                                                             //成功状态吗
	RES_FAIL  = -1                                                            //失败状态吗
	REGULAR   = "^((13[0-9])|(14[5|7])|(15([0-3]|[5-9]))|(18[0,5-9]))\\d{8}$" //手机号码正则
)

//api return
func Fail(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, gin.H{"msg": err, "status": "FAIL", "res": RES_FAIL})

}

func OK(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"msg": msg, "status": "OK", "res": RES_OK})
}

func OKWithData(c *gin.Context, msg string, data map[string]interface{}) {
	c.JSON(http.StatusOK, gin.H{"msg": msg, "status": "OK", "res": RES_OK, "data": data})
}

//sha256
func GetSha256Code(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Md5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

//random 随机数
func RandCode(number int) string {
	var result string
	for i := 0; i < number; i++ {
		result = result + strconv.Itoa(rand.Intn(9))
	}
	return result
}

func ValidatePhone(mobileNum string) bool {
	reg := regexp.MustCompile(REGULAR)
	return reg.MatchString(mobileNum)
}

func SecrectKey(key string) string {
	return Md5(key)
}
