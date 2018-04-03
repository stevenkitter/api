package v1

//微信模块 负责跟微信交互 存入redis关键数据
import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"github.com/stevenkitter/api/v1/common"
	"github.com/stevenkitter/api/v1/xhttp"

	"github.com/gin-gonic/gin"
)

const (
	//Token is
	Token          = ""
	//AppID is
	AppID          = ""
	//AppSecrect is
	AppSecrect     = ""
	//EncodingAESKey is
	EncodingAESKey = "7WfXuJfsGHYqt5eSPH8Gg7B9Y115vU8dx4Z48rZbzH1"
)
//WxHandler is
/**
 * [微信10分钟走一次这个接口，所有数据都是加密的]
 * @type {[post请求]}
 */
func WxHandler(c *gin.Context) {
	//签名
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	encryptType := c.Query("encrypt_type")
	msgSignature := c.Query("msg_signature")
	if encryptType != "aes" {
		log.Printf("非法请求过来了")
		success(c)
		return
	}
	//验证签名
	var encMessage EncMessage
	if err := c.ShouldBind(&encMessage); err == nil {
		if !checkSignature(Token, timestamp, nonce, encMessage.Encrypt, msgSignature) {
			log.Printf("非法请求过来了")
			success(c)
			return
		}
		//解密数据到模型
		EncodingAESKeyBase64, _ := base64.StdEncoding.DecodeString(EncodingAESKey + "=")
		res, err := DecryptMsg(encMessage.Encrypt, EncodingAESKeyBase64, AppID)
		if err != nil {
			log.Printf(err.Error())
			success(c)
			return
		}
		receivedMessage := &ReceiveMessage{}
		xmlerr := xml.Unmarshal(res, receivedMessage)
		if xmlerr != nil {
			log.Printf("err is %s", xmlerr.Error())
			success(c)
			return
		}
		//获得票据 并存入redis
		if receivedMessage.InfoType == "component_verify_ticket" {
			// log.Info(receivedMessage.ComponentVerifyTicket)

			reErr := common.RedisSaveString("component_verify_ticket", receivedMessage.ComponentVerifyTicket)
			if reErr != nil {
				log.Printf("err is %s", reErr.Error())
				success(c)
				return
			}

		} else {
			mysql, err := common.Mariadb()
			defer mysql.Close()
			if err != nil {
				log.Printf("mysql wrong %s", err.Error())
				success(c)
			}
			sql := "REPLACE INTO wxAuthInfo(authorize_info, authorization_code, " +
				"authorizer_appid, authorization_code_expiredTime, pre_auth_code, " +
				"create_time, client_id)" +
				" VALUES (?, ?, ?, ?, ?, ? (SELECT jl_userID FROM JL_User WHERE token = ?))"
			//授权取消 完成 重新授权等 存储一下信息
			_, err = mysql.Exec(sql, receivedMessage.InfoType, receivedMessage.AuthorizationCode,
				receivedMessage.AuthorizerAppid, receivedMessage.AuthorizationCodeExpiredTime,
				receivedMessage.PreAuthCode, receivedMessage.CreateTime, c.GetHeader("token"))
			if err != nil {
				log.Printf("mysql wrong %s", err.Error())
				success(c)
			}
		}
		apiComponentToken, _ := APIComponentToken()
		log.Println(apiComponentToken)
		success(c)
	} else {
		log.Printf(err.Error())
		success(c)
		return
	}
}
//GetPreAuthCode is
/**
 * 暴露给接口使用的
 * @type {预授权码}
 */
func GetPreAuthCode() (string, error) {
	code, err := preAuthCode()
	return code, err
}
func success(c *gin.Context) {
	c.String(200, "success")
}

//APIComponentToken 获取token
func APIComponentToken() (string, error) {
	//redis查询
	componentAccessToken, err := common.RedisGETString("component_access_token")
	if componentAccessToken != "" { //过期了实效了
		return componentAccessToken, nil
		
	}
	log.Printf("redis error is %s", err.Error())
	componentAccessToken, err = requestAPIComponentToken()
	return componentAccessToken, err
}
//requestAPIComponentToken is
func requestAPIComponentToken() (string, error) {
	ticket, err := common.RedisGETString("component_verify_ticket")
	if ticket != "" {
		return ticket, err
	}
	request := APIComponentTokenRequest{AppID, AppSecrect, ticket}
	url := "https://api.weixin.qq.com/cgi-bin/component/api_component_token"
	body, err := xhttp.PostStruct(url, request)
	if err != nil {
		return "", err
	}
	var response APIComponentTokenResponse
	err = json.Unmarshal(body, &response) //json 转 对象
	if err != nil {
		return "", err
	}
	//请求到数据 存一下redis
	err = common.RedisSaveStringEx("component_access_token", response.ComponentAccessToken, strconv.Itoa(response.ExpiresIn))
	if err != nil {
		log.Printf(err.Error()) //存储有问题
	}

	return response.ComponentAccessToken, nil
}

//post
func checkSignature(token, timestamp, nonce, encrypt, sign string) bool {
	tmpArr := []string{token, timestamp, nonce, encrypt}
	sort.Strings(tmpArr)
	tmpStr := strings.Join(tmpArr, "")
	actual := fmt.Sprintf("%x", sha1.Sum([]byte(tmpStr)))
	return actual == sign
}

//preAuthCode 用户需要授权 获取这个字段
func preAuthCode() (string, error) {
	preAuthCode, err := common.RedisGETString("pre_auth_code")
	if preAuthCode != "" { //过期了实效了
		return preAuthCode, nil
	}
	log.Printf("redis error is %s", err.Error())
	preAuthCode, err = requestPreAuthCode()
	return preAuthCode, err
}

//
func requestPreAuthCode() (string, error) {
	post := map[string]interface{}{"component_appid": AppID}
	thisJSON, _ := common.MapToJson(post)
	token, err := APIComponentToken()
	if err != nil {
		log.Printf("token err %s", err.Error())
		return "", err
	}
	url := "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token=" +
		token
	body, err := xhttp.Post(url, []byte(thisJSON))
	if err != nil {
		log.Printf("url(%s) request err(%s)", url, err.Error())
		return "", err
	}
	var response PreAuthCode
	err = json.Unmarshal(body, &response) //json 转 对象
	if err != nil {
		return "", err
	}
	//请求到数据 存一下redis
	err = common.RedisSaveStringEx("pre_auth_code", response.PreAuthCode, strconv.Itoa(response.ExpiresIn))
	if err != nil {
		log.Printf(err.Error()) //存储有问题
		return "", err
	}

	return response.PreAuthCode, nil

}
