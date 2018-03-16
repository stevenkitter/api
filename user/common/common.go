package common

import (
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"strconv"
)

//api return
func Fail(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, gin.H{"msg": err, "status": "FAIL", "res": -1})

}

func OK(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"msg": msg, "status": "OK", "res": 0})
}

func OKWithData(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"msg": msg, "status": "OK", "res": 0, "data": data})
}

//sha256
func GetSha256Code(s string) string {
	h := sha256.New()
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
