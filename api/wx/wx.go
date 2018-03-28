package wx

//微信模块 负责跟微信交互 存入redis关键数据
import (
	"api/common"
	"api/xhttp"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	"sort"
	"strconv"
	"strings"
)
import log "github.com/cihub/seelog"

const (
	Token          = ""
	AppId          = ""
	AppSecrect     = ""
	EncodingAESKey = "7WfXuJfsGHYqt5eSPH8Gg7B9Y115vU8dx4Z48rZbzH1"
)

func WxHandler(c *gin.Context) {
	//签名
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	encrypt_type := c.Query("encrypt_type")
	msg_signature := c.Query("msg_signature")
	if encrypt_type != "aes" {
		log.Error("非法请求过来了")
		success(c)
		return
	}

	var encMessage EncMessage
	if err := c.ShouldBind(&encMessage); err == nil {
		if !checkSignature(Token, timestamp, nonce, encMessage.Encrypt, msg_signature) {
			log.Error("非法请求过来了")
			success(c)
			return
		}
		EncodingAESKeyBase64, _ := base64.StdEncoding.DecodeString(EncodingAESKey + "=")
		res, err := DecryptMsg(encMessage.Encrypt, EncodingAESKeyBase64, AppId)
		if err != nil {
			log.Error(err.Error())
			success(c)
			return
		}
		receivedMessage := &ReceiveMessage{}
		xmlerr := xml.Unmarshal(res, receivedMessage)
		if xmlerr != nil {
			log.Error("err is %s", xmlerr.Error())
			success(c)
			return
		}
		log.Info(receivedMessage.ComponentVerifyTicket)
		//每10分钟存一下ticket
		re_err := common.RedisSaveString("component_verify_ticket", receivedMessage.ComponentVerifyTicket)
		if re_err != nil {
			log.Error("err is %s", re_err.Error())
			success(c)
			return
		}
		//看看token有效没 无效就刷新
		api_component_token, _ := api_component_token()
		log.Info(api_component_token)
		success(c)
	} else {
		log.Error(err.Error())
		success(c)
		return
	}
}

func success(c *gin.Context) {
	c.String(200, "success")
}

// 获取token
func api_component_token() (string, error) {
	//redis查询
	component_access_token, err := common.RedisGETString("component_access_token")
	if component_access_token == "" { //过期了实效了
		log.Error("redis error is %s", err.Error())
		component_access_token, err = request_api_component_token()
		return component_access_token, err
	} else {
		return component_access_token, nil
	}
}

func request_api_component_token() (string, error) {
	ticket, err := common.RedisGETString("component_verify_ticket")
	if err != nil {
		return "", err
	}
	request := ApiComponentTokenRequest{AppId, AppSecrect, ticket}
	url := "https://api.weixin.qq.com/cgi-bin/component/api_component_token"
	body, err := xhttp.PostStruct(url, request)
	if err != nil {
		return "", err
	}
	var response ApiComponentTokenResponse
	err = json.Unmarshal(body, &response) //json 转 对象
	if err != nil {
		return "", err
	}
	//请求到数据 存一下redis
	err = common.RedisSaveStringEx("component_access_token", response.Component_access_token, strconv.Itoa(response.Expires_in))
	if err != nil {
		log.Error(err.Error()) //存储有问题
	}

	return response.Component_access_token, nil
}

//post
func checkSignature(token, timestamp, nonce, encrypt, sign string) bool {
	tmpArr := []string{token, timestamp, nonce, encrypt}
	sort.Strings(tmpArr)
	tmpStr := strings.Join(tmpArr, "")
	actual := fmt.Sprintf("%x", sha1.Sum([]byte(tmpStr)))
	return actual == sign
}

//用户需要授权 获取这个字段
func pre_auth_code() (string, error) {
	pre_auth_code, err := common.RedisGETString("pre_auth_code")
	if pre_auth_code == "" { //过期了实效了
		log.Error("redis error is %s", err.Error())
		pre_auth_code, err = request_pre_auth_code()
		return pre_auth_code, err
	} else {
		return pre_auth_code, nil
	}
}

//
func request_pre_auth_code() (string, error) {
	post := map[string]interface{}{"component_appid": AppId}
	thisJson, _ := common.MapToJson(post)
	token, err := api_component_token()
	if err != nil {
		log.Error("token err %s", err.Error())
		return "", err
	}
	url := "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token=" +
		token
	body, err := xhttp.Post(url, []byte(thisJson))
	if err != nil {
		log.Error("url(%s) request err(%s)", url, err.Error())
		return "", err
	}
	var response Pre_auth_code
	err = json.Unmarshal(body, &response) //json 转 对象
	if err != nil {
		return "", err
	}
	//请求到数据 存一下redis
	err = common.RedisSaveStringEx("pre_auth_code", response.Pre_auth_code, strconv.Itoa(response.Expires_in))
	if err != nil {
		log.Error(err.Error()) //存储有问题
		return "", err
	}

	return response.Pre_auth_code, nil

}
