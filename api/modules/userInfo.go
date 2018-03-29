package modules

import (
	"api/common"
	"api/wx"
	"api/xhttp"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
)

/**
 * 用户登陆的基本信息 授权前什么信息也没有 授权后展示微信所给的信息
 * 这个信息 平台的数据库也保存一份 方便以后客户使用
 * 每次登陆 平台都主动帮组刷新这个数据
 */
func UserInfo(c *gin.Context) {
	token, err := wx.Api_component_token()
	if token == "" {
		common.Fail(c, err.Error())
		return
	}
}

/**
 * 请求预授权码 用来换取用户的授权码
 * 当用户点击了平台的授权按钮 请求这个数据
 * 平台必须存一下这个code 过期了就重新请求（客户端时间为8分钟）
 * redis缓存10分钟时间 不要重复请求
 * 重复请求也没关系 存在redis缓存里呢
 */
func RequestForPreAuthCode(c *gin.Context) {
	code, err := wx.Get_pre_auth_code()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKWithData(c, "", map[string]interface{}{"pre_auth_code": code})
}

/**
 * 客户端在接受到微信回调时，需要通过auth_code来换取客户凭证
 * 这块的信息很重要需要妥善存放
 * 客户端post这个数据过来
 * @params component_appid, authorization_code
 */
func RequestAuthInfo(c *gin.Context) {
	token, err := wx.Api_component_token()
	if token == "" {
		common.Fail(c, err.Error())
		return
	}
	var request RequestAuthInfo_Request_Model
	if err := c.ShouldBind(&request); err != nil {
		common.Fail(c, err.Error())
		return
	}
	url := "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token=" +
		token
	responseByte, err := xhttp.PostStruct(url, request)

	var response RequestAuthInfo_Response_Model
	err = json.Unmarshal(responseByte, &response)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	//处理一下数据给平台 并存储一下数据 挺麻烦这块
	funcInfos := response.Authorization_info.Func_info
	for _, category := range funcInfos {
		category.Funcscope_category.Name = categoryName(category.Funcscope_category.Id)
	}
	//数据已就位 赶紧存储
	mysql, err := common.Mariadb()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	defer mysql.Close()

	client_token := c.Request.Header.Get("token")
	// sql := "SELECT jl_userID FROM JL_User WHERE token = ?"
	// var userId string
	// mysql.QueryRow(sql, client_token).Scan(&userId)
	sql := "UPDATE JL_User SET authorizer_appid = ? WHERE token = ?"
	_, err = mysql.Exec(sql, response.Authorization_info.Authorizer_appid, client_token)
	// 表里已经记录这个用户的小程序id
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	sql = "REPLACE INTO wxAuthInfo(authorizer_appid, authorizer_access_token, " +
		"expires_in, authorizer_refresh_token, client_id)" +
		" VALUES (?, ?, ?, ?, (SELECT jl_userID FROM JL_User WHERE token = ?))"
	_, err = mysql.Exec(sql, response.Authorization_info.Authorizer_appid,
		response.Authorization_info.Authorizer_access_token,
		response.Authorization_info.Expires_in,
		response.Authorization_info.Authorizer_refresh_token,
		client_token)
	//我日 这么长的sql语句
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	sql = "REPLACE INTO wxAuthFuncInfo(func_id, func_name, authorizer_appid) VALUES "
	paramstrings := make([]string, len(response.Authorization_info.Func_info))
	params := make([]interface{}, 0)
	for index, value := range response.Authorization_info.Func_info {
		paramstrings[index] = "(?, ?, ?)"
		params = append(params, value.Funcscope_category.Id)
		params = append(params, value.Funcscope_category.Name)
		params = append(params, response.Authorization_info.Authorizer_appid)
	}
	paramstring := strings.Join(paramstrings, ", ")
	sql += paramstring
	_, err = mysql.Exec(sql, params)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	//redis 放入authorizer_access_token
	err = common.RedisSaveStringEx(response.Authorization_info.Authorizer_appid+"authorizer_access_token", response.Authorization_info.Authorizer_access_token, "7000")
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, "已授权成功")
}

/**
 * 根据用户的token获取authorizer_access_token来处理用户的权限
 * @params token
 */
func Authorizer_access_token(token string) (string, error) {
	mysql, err := common.Mariadb()
	defer mysql.Close()
	if err != nil {
		return "", err
	}
	sql := "SELECT authorizer_appid, authorizer_refresh_token FROM wxAuthInfo " +
		"WHERE client_id = (SELECT jl_userID FROM jl_userID WHERE token = ?)"
	var request RequestAuthorizer_Access_Token_Request_Model
	request.Component_appid = wx.AppId
	mysql.QueryRow(sql, token).Scan(&request.Authorizer_appid, &request.Authorizer_refresh_token)
	ac_token, err := common.RedisGETString(request.Authorizer_appid + "authorizer_access_token")
	if err == nil {
		return ac_token, nil
	}
	com_token, err := wx.Api_component_token()
	if com_token == "" {
		return "", err
	}
	url := "https:// api.weixin.qq.com /cgi-bin/component/api_authorizer_token?component_access_token=" +
		com_token
	var response RequestAuthorizer_Access_Token_Response_Model
	err = xhttp.PostModel(url, request, *response)
	if err != nil {
		return "", err
	}
	//存一下数据token
	err = common.RedisSaveStringEx(request.Authorizer_appid+"authorizer_access_token", response.Authorizer_access_token, "7000")
	if err != nil {
		return "", err
	}
	return response.Authorizer_access_token, nil
}

func categoryName(id int) string {
	switch id {
	case 1:
		return "消息管理权限"
	case 2:
		return "用户管理权限"
	case 3:
		return "账号服务权限"
	case 4:
		return "网页服务权限"
	case 5:
		return "微信小店权限"
	case 6:
		return "微信多客服权限"
	case 7:
		return "群发与通知权限"
	case 8:
		return "微信卡券权限"
	case 9:
		return "微信扫一扫权限"
	case 10:
		return "微信连WIFI权限"
	case 11:
		return "素材管理权限"
	case 12:
		return "微信摇周边权限"
	case 13:
		return "微信门店权限"
	case 14:
		return "微信支付权限"
	case 15:
		return "自定义菜单权限"
	case 16:
		return "获取认证状态及信息"
	case 17:
		return "帐号管理权限（小程序）"
	case 18:
		return "开发管理与数据分析权限（小程序）"
	case 19:
		return "客服消息管理权限（小程序）"
	case 20:
		return "微信登录权限（小程序）"
	case 21:
		return "数据分析权限（小程序）"
	case 22:
		return "城市服务接口权限"
	case 23:
		return "广告管理权限"
	case 24:
		return "开放平台帐号管理权限"
	case 25:
		return "开放平台帐号管理权限（小程序）"
	case 26:
		return "微信电子发票权限"

	default:
		return ""
	}
	return ""
}

//-----------------------------------------------------------------------------
/**
 * 接口所有使用的模型
 */

type RequestAuthInfo_Request_Model struct {
	Component_appid    string `json:"component_appid"`
	Authorization_code string `json:"authorization_code"`
}
type RequestAuthInfo_Response_Model struct {
	Authorization_info Authorization_info_Model `json:"authorization_info"`
}
type Authorization_info_Model struct {
	Authorizer_appid         string            `json:"authorizer_appid"`
	Authorizer_access_token  string            `json:"authorizer_access_token"`
	Expires_in               int               `json:"expires_in"`
	Authorizer_refresh_token string            `json:"authorizer_refresh_token"`
	Func_info                []Func_Info_Model `json:"func_info"`
}
type Func_Info_Model struct {
	Funcscope_category Funcscope_Category_Model `json:"funcscope_category"`
}
type Funcscope_Category_Model struct {
	Id   int `json:"id"`
	Name string
}

//------------------------------
type RequestAuthorizer_Access_Token_Request_Model struct {
	Component_appid          string `json:"component_appid"`
	Authorizer_appid         string `json:"authorizer_appid"`
	Authorizer_refresh_token string `json:"authorizer_refresh_token"`
}
type RequestAuthorizer_Access_Token_Response_Model struct {
	Authorizer_access_token  string `json:"authorizer_access_token"`
	Expires_in               string `json:"expires_in"`
	Authorizer_refresh_token string `json:"authorizer_refresh_token"`
}
