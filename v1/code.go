package v1

import (
	"strconv"
	"time"
	"github.com/stevenkitter/api/v1/common"
	"github.com/stevenkitter/api/v1/xhttp"

	"github.com/gin-gonic/gin"
)

//常用参数
const (
	sdkAppid    = "1400063858"                       //appid
	sdkAppKey   = "345fa4781613c887768ff529cc782b57" //appkey
	tag         = 79442                              //模版id
	china       = "86"                               //电话前缀
	randomCount = 8                                  //8位随机数
)

//TelType is
type TelType struct {
	Mobile     string `form:"mobile" json:"mobile" binding:"required"`
	Nationcode string `form:"nationcode" json:"nationcode" binding:"required"`
}

//CodeParamType 腾讯云短信api需要的参数
type CodeParamType struct {
	Params []string `form:"phone" json:"params" binding:"required"`
	Sig    string   `form:"sig" json:"sig" binding:"required"`
	Tel    TelType  `form:"tel" json:"tel" binding:"required"`
	Time   int64    `form:"time" json:"time" binding:"required"`
	TplID int      `form:"tpl_id" json:"tpl_id" binding:"required"`
}

//Respon api返回格式
type Respon struct {
	Result int    `json:"result"`
	Errmsg string `json:"errmsg"`
	Ext    string `json:"ext"`
	Fee    int    `json:"fee"`
	Sid    string `json:"sid"`
}

//GetCode 获取验证码 采用腾讯云的短信api
func GetCode(c *gin.Context) {

	phone := c.Query("phone")

	if phone == "" {
		common.Fail(c, "缺少手机号码参数")
		return
	}

	if !common.ValidatePhone(phone) {
		common.Fail(c, "错误的手机号码格式")
		return
	}
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

	sixCo := sixCode()
	var param = CodeParamType{
		Params: []string{sixCo},
		Sig:    sig,
		Tel:    tel,
		Time:   timestamp,
		TplID: tag,
	} //post数据

	// jsonStr, err := json.Marshal(param) //json
	//
	// if err != nil {
	// 	common.Fail(c, err.Error())
	// 	return
	// }
	// body, err := xhttp.Post(url, jsonStr) //传送json 返回 []byte
	// if err != nil {
	// 	common.Fail(c, err.Error())
	// 	return
	// }
	// var respond Respon
	// jsonErr := json.Unmarshal(body, &respond) //json 转 对象
	// if jsonErr != nil {
	// 	common.Fail(c, jsonErr.Error())
	// 	return
	// }
	respond := &Respon{}
	err := xhttp.PostModel(url, param, respond)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	if respond.Result != 0 {
		common.Fail(c, respond.Errmsg)
		return
		
	} 
	//存一下redis 以后接口匹配要用
	redis, err := common.Redis()
	defer redis.Close()
	if err != nil {
		common.Fail(c, err.Error()) //链接redis失败
		return
	}
	reName := phone + "code"
	_, err = redis.Do("SET", reName, sixCo, "EX", "300")
	if err != nil {
		common.Fail(c, err.Error()) //保存数据失败
		return
	}
	common.OK(c, "已发送到手机") //返回<->
}

//验证码 6位的
func sixCode() string {
	return common.RandCode(6)
}
