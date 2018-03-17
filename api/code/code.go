package code

import (
	"api/common"
	"api/xhttp"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

//常用参数
const (
	sdkAppid    = "1400063858"                       //appid
	sdkAppKey   = "345fa4781613c887768ff529cc782b57" //appkey
	tag         = 79442                              //模版id
	china       = "86"                               //电话前缀
	randomCount = 8                                  //8位随机数
)

//电话格式
type TelType struct {
	Mobile     string `form:"mobile" json:"mobile" binding:"required"`
	Nationcode string `form:"nationcode" json:"nationcode" binding:"required"`
}

//腾讯云短信api需要的参数
type CodeParamType struct {
	Params []string `form:"phone" json:"params" binding:"required"`
	Sig    string   `form:"sig" json:"sig" binding:"required"`
	Tel    TelType  `form:"tel" json:"tel" binding:"required"`
	Time   int64    `form:"time" json:"time" binding:"required"`
	Tpl_id int      `form:"tpl_id" json:"tpl_id" binding:"required"`
}

//api返回格式
type Respon struct {
	Result int    `json:"result"`
	Errmsg string `json:"errmsg"`
	Ext    string `json:"ext"`
	Fee    int    `json:"fee"`
	Sid    string `json:"sid"`
}

//获取验证码 采用腾讯云的短信api
func GetCode(c *gin.Context) {

	phone := c.Query("phone")

	ranCo := common.RandCode(randomCount) //8位随机数

	url := "https://yun.tim.qq.com/v5/tlssmssvr/sendsms?sdkappid=" +
		sdkAppid + "&random=" + ranCo

	t := time.Now()
	timestamp := t.Unix()

	timeStr := strconv.FormatInt(timestamp, 10) //事件戳字符串

	sigStr := "appkey=" + sdkAppKey + "&random=" +
		ranCo + "&time=" + timeStr + "&mobile=" + phone

	sig := common.GetSha256Code(sigStr) //签名

	var tel TelType
	tel.Mobile = phone
	tel.Nationcode = china //号码格式 中国的号码

	var param CodeParamType //post数据
	param.Params = []string{sixCode()}
	param.Sig = sig
	param.Tel = tel
	param.Time = timestamp
	param.Tpl_id = tag

	jsonStr, err := json.Marshal(param) //json

	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	body, err := xhttp.Post(url, jsonStr) //传送json 返回 []byte
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	var respond Respon
	jsonErr := json.Unmarshal(body, &respond) //json 转 对象
	if jsonErr != nil {
		common.Fail(c, jsonErr.Error())
		return
	}
	if respond.Result == 0 {
		common.OK(c, "已发送到手机") //返回<->
		return
	} else {
		common.Fail(c, respond.Errmsg)
		return
	}

}

//验证码 6位的
func sixCode() string {
	return common.RandCode(6)
}
