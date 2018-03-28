package wx

import (
	"encoding/xml"
)

//微信加密消息固定格式
type EncMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"-"` // 开发者微信号
	Encrypt      string   // 加密的消息报文
	MsgSignature string   // 报文签名
	TimeStamp    string   // 时间戳
	Nonce        string   // 随机数
}

//微信每隔10分钟发过来一个数据ticket，需要解密
type ReceiveMessage struct {
	XMLName               xml.Name `xml:"xml"`
	AppId                 string   `xml:"AppId"`                 //第三方平台appid
	CreateTime            string   `xml:"CreateTime"`            //时间戳
	InfoType              int64    `xml:"InfoType"`              //component_verify_ticket
	ComponentVerifyTicket string   `xml:"ComponentVerifyTicket"` //Ticket内容
}

type ApiComponentTokenRequest struct {
	Component_appid         string `json:"component_appid"`
	Component_appsecret     string `json:"component_appsecret"`
	Component_verify_ticket string `json:"component_verify_ticket"`
}

type ApiComponentTokenResponse struct {
	Component_access_token string `json:"component_access_token"`
	Expires_in             int    `json:"expires_in"`
}

type Pre_auth_code struct {
	Pre_auth_code string `json:"pre_auth_code"`
	Expires_in    int    `json:"expires_in"`
}
