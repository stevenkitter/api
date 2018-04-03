package v1

import (
	"encoding/json"
	"strings"
	"github.com/stevenkitter/api/v1/common"
	"github.com/stevenkitter/api/v1/xhttp"

	"github.com/gin-gonic/gin"
)

//UserInfo public 暴露给接口的---------------------------****-----------

//UserInfo 用户登陆的基本信息 授权前什么信息也没有 授权后展示微信所给的信息
 /* 这个信息 平台的数据库也保存一份 方便以后客户使用
 * 每次登陆 平台都主动帮组刷新这个数据
 * GET
 * @params nil
 */
func UserInfo(c *gin.Context) {
	token, err := APIComponentToken()
	if token == "" {
		common.Fail(c, err.Error())
		return
	}
	url := "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_info?component_access_token=" +
		token
	response := &RequestUserInfoResponseModel{}
	clientToken := c.Request.Header.Get("token") //平台使用的token
	auAppid, err := AuthorizerAppid(clientToken)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	obj := RequestUserInfoRequestModel{AppID, auAppid}

	err = xhttp.PostModel(url, obj, response)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	data := common.StructToMap(response)
	common.OKWithData(c, "success", data)
}

//RequestForPreAuthCode is
 /* 请求预授权码 用来换取用户的授权码
 * 当用户点击了平台的授权按钮 请求这个数据
 * 平台必须存一下这个code 过期了就重新请求（客户端时间为8分钟）
 * redis缓存10分钟时间 不要重复请求
 * 重复请求也没关系 存在redis缓存里呢
 * @params nil
 */
func RequestForPreAuthCode(c *gin.Context) {
	code, err := GetPreAuthCode()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKWithData(c, "", map[string]interface{}{"pre_auth_code": code})
}

//RequestAuthInfo is
/**
 * 客户端在接受到微信回调时，需要通过auth_code来换取客户凭证
 * 这块的信息很重要需要妥善存放
 * 客户端post这个数据过来
 * POST
 * @params component_appid, authorization_code
 */
func RequestAuthInfo(c *gin.Context) {
	token, err := APIComponentToken()
	if token == "" {
		common.Fail(c, err.Error())
		return
	}
	var request RequestAuthInfoRequestModel
	if err := c.ShouldBindJSON(&request); err != nil {
		common.Fail(c, err.Error())
		return
	}
	url := "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token=" +
		token
	responseByte, err := xhttp.PostStruct(url, request)

	var response RequestAuthInfoResponseModel
	err = json.Unmarshal(responseByte, &response)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	//处理一下数据给平台 并存储一下数据 挺麻烦这块
	funcInfos := response.AuthorizationInfo.FuncInfo
	for _, category := range funcInfos {
		category.FuncscopeCategory.Name = categoryName(category.FuncscopeCategory.ID)
	}
	//数据已就位 赶紧存储
	mysql, err := common.Mariadb()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	defer mysql.Close()

	clientToken := c.Request.Header.Get("token")
	// sql := "SELECT jl_userID FROM JL_User WHERE token = ?"
	// var userId string
	// mysql.QueryRow(sql, client_token).Scan(&userId)
	sql := "UPDATE JL_User SET authorizer_appid = ? WHERE token = ?"
	_, err = mysql.Exec(sql, response.AuthorizationInfo.AuthorizerAppid, clientToken)
	// 表里已经记录这个用户的小程序id
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	sql = "REPLACE INTO wxAuthInfo(authorizer_appid, authorizer_access_token, " +
		"expires_in, authorizer_refresh_token, client_id)" +
		" VALUES (?, ?, ?, ?, (SELECT jl_userID FROM JL_User WHERE token = ?))"
	_, err = mysql.Exec(sql, response.AuthorizationInfo.AuthorizerAppid,
		response.AuthorizationInfo.AuthorizerAccessToken,
		response.AuthorizationInfo.ExpiresIn,
		response.AuthorizationInfo.AuthorizerRefreshToken,
		clientToken)
	//我日 这么长的sql语句
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	sql = "REPLACE INTO wxAuthFuncInfo(func_id, func_name, authorizer_appid) VALUES "
	paramstrings := make([]string, len(response.AuthorizationInfo.FuncInfo))
	params := make([]interface{}, 0)
	for index, value := range response.AuthorizationInfo.FuncInfo {
		paramstrings[index] = "(?, ?, ?)"
		params = append(params, value.FuncscopeCategory.ID)
		params = append(params, value.FuncscopeCategory.Name)
		params = append(params, response.AuthorizationInfo.AuthorizerAppid)
	}
	paramstring := strings.Join(paramstrings, ", ")
	sql += paramstring
	_, err = mysql.Exec(sql, params)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	//redis 放入authorizer_access_token
	err = common.RedisSaveStringEx(response.AuthorizationInfo.AuthorizerAppid+"authorizer_access_token", response.AuthorizationInfo.AuthorizerAccessToken, "7000")
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, "已授权成功")
}
//AuthorizerOption is
/**
 * 授权方的选项信息
 * location_report 地理位置上报选项 0无 1会话时上报 2每5s上报
 * voice_recognize 0关闭 1打开
 * customer_service 0关闭 1打开
 * GET
 * @params option_name
 */
func AuthorizerOption(c *gin.Context) {
	optionName := c.Query("option_name")
	clientToken := c.GetHeader("token")
	authAppid, err := AuthorizerAppid(clientToken)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	compToken, err := APIComponentToken()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	url := "https://api.weixin.qq.com/cgi-bin/component/ api_get_authorizer_option?component_access_token=" +
	compToken
	obj := new(AuthorizerOptionRequestModel)
	obj.ComponentAppid = AppID
	obj.AuthorizerAppid = authAppid
	obj.OptionName = optionName
	response := &AuthorizerOptionReponseModel{}
	data, err := xhttp.PostModelToData(url, obj, response)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKData(c, data)
}
//SetAuthorizerOption is
/**
 * POST
 * @params option_name option_value
 */
func SetAuthorizerOption(c *gin.Context) {
	var request AuthorizerOptionSETRequestModel
	if err := c.ShouldBindJSON(&request); err != nil {
		common.Fail(c, err.Error())
		return
	}
	authAppid, err := AuthorizerAppid(c.GetHeader("token"))
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	request.ComponentAppid = AppID
	request.AuthorizerAppid = authAppid
	compToken, err := APIComponentToken()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	url := "https://api.weixin.qq.com/cgi-bin/component/ api_set_authorizer_option?component_access_token=xxxx" +
	compToken
	response := &WXResponse{}
	data, err := xhttp.PostModelToData(url, request, response)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKWithData(c, "设置成功！", data)
}
//AuthorizerAccessToken is
/**
 * internal 暴露给内部使用的---------------------------****-----------
 */
/**
 * 根据用户的token获取authorizer_access_token来处理用户的权限
 * @params token
 */
func AuthorizerAccessToken(token string) (string, error) {
	mysql, err := common.Mariadb()
	defer mysql.Close()
	if err != nil {
		return "", err
	}
	sql := "SELECT authorizer_appid, authorizer_refresh_token FROM wxAuthInfo " +
		"WHERE client_id = (SELECT jl_userID FROM jl_userID WHERE token = ?)"
	var request RequestAuthorizerAccessTokenRequestModel
	request.ComponentAppid = AppID
	mysql.QueryRow(sql, token).Scan(&request.AuthorizerAppid, &request.AuthorizerRefreshToken)
	acToken, err := common.RedisGETString(request.AuthorizerAppid + "authorizer_access_token")
	if err == nil {
		return acToken, nil
	}
	comToken, err := APIComponentToken()
	if comToken == "" {
		return "", err
	}
	url := "https:// api.weixin.qq.com /cgi-bin/component/api_authorizer_token?component_access_token=" +
	comToken
	response := &RequestAuthorizerAccessTokenResponseModel{}
	err = xhttp.PostModel(url, request, response)
	if err != nil {
		return "", err
	}
	//存一下数据token
	err = common.RedisSaveStringEx(request.AuthorizerAppid+"authorizer_access_token", response.AuthorizerAccessToken, "7000")
	if err != nil {
		return "", err
	}
	return response.AuthorizerAccessToken, nil
}
//AuthorizerAppid IS
func AuthorizerAppid(token string) (string, error) {
	mysql, err := common.Mariadb()
	defer mysql.Close()
	if err != nil {
		return "", err
	}
	sql := "SELECT authorizer_appid FROM wxAuthInfo WHERE client_id = " +
		"(SELECT jl_userID FROM JL_User WHERE token = ?)"
	var authorizerAppid string
	err = mysql.QueryRow(sql, token).Scan(&authorizerAppid)
	if err != nil {
		return "", err
	}
	return authorizerAppid, nil
}

/**
 * private 私有的---------------------------****-----------
 */
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
}

//RequestAuthInfoRequestModel IS
type RequestAuthInfoRequestModel struct {
	ComponentAppid    string `json:"component_appid"`
	AuthorizationCode string `json:"authorization_code"`
}
//RequestAuthInfoResponseModel IS
type RequestAuthInfoResponseModel struct {
	AuthorizationInfo AuthorizationInfoModel `json:"authorization_info"`
}
//AuthorizationInfoModel IS
type AuthorizationInfoModel struct {
	AuthorizerAppid         string            `json:"authorizer_appid"`
	AuthorizerAccessToken  string            `json:"authorizer_access_token"`
	ExpiresIn               int               `json:"expires_in"`
	AuthorizerRefreshToken string            `json:"authorizer_refresh_token"`
	FuncInfo                []FuncInfoModel `json:"func_info"`
}
//FuncInfoModel IS
type FuncInfoModel struct {
	FuncscopeCategory FuncscopeCategoryModel `json:"funcscope_category"`
}
//FuncscopeCategoryModel IS
type FuncscopeCategoryModel struct {
	ID   int `json:"id"`
	Name string
}

// RequestAuthorizerAccessTokenRequestModel ------------------------------
type RequestAuthorizerAccessTokenRequestModel struct {
	ComponentAppid          string `json:"component_appid"`
	AuthorizerAppid         string `json:"authorizer_appid"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}
//RequestAuthorizerAccessTokenResponseModel IS
type RequestAuthorizerAccessTokenResponseModel struct {
	AuthorizerAccessToken  string `json:"authorizer_access_token"`
	ExpiresIn               string `json:"expires_in"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

//RequestUserInfoRequestModel ----------------------------
type RequestUserInfoRequestModel struct {
	ComponentAppid  string `json:"component_appid"`
	AuthorizerAppid string `json:"authorizer_appid"`
}
//RequestUserInfoResponseModel IS
type RequestUserInfoResponseModel struct {
	AuthorizerInfo    AuthorizerInfoModel    `json:"authorizer_info"`
	AuthorizationInfo AuthorizationInfoModel `json:"authorization_info"`
}
//AuthorizerInfoModel IS
type AuthorizerInfoModel struct {
	NickName         string `json:"nick_name"`
	HeadImg          string `json:"head_img"`
	ServiceTypeInfo struct {
		ID int `json:"id"`
	} `json:"service_type_info"`
	VerifyTypeInfo struct {
		ID int `json:"id"`
	} `json:"verify_type_info"`
	UserName      string `json:"user_name"` //12
	PrincipalName string `json:"principal_name"`
	BusinessInfo  struct {
		OpenStore int `json:"open_store"`
		OpenScan  int `json:"open_scan"`
		OpenPay   int `json:"open_pay"`
		OpenCard  int `json:"open_card"`
		OpenShake int `json:"open_shake"`
	} `json:"business_info"`
	Alias           string `json:"alias"` //公众号
	QrcodeURL       string `json:"qrcode_url"`
	Signature       string `json:"signature"` //小程序
	MiniProgramInfo struct {
		Network struct {
			RequestDomain   []string `json:"RequestDomain"`
			WsRequestDomain []string `json:"WsRequestDomain"`
			UploadDomain    []string `json:"UploadDomain"`
			DownloadDomain  []string `json:"DownloadDomain"`
		} `json:"network"`
		Categories []struct {
			First  string `json:"first"`
			Second string `json:"second"`
		} `json:"categories"`
		VisitStatus int `json:"visit_status"`
	} `json:"MiniProgramInfo"`
}

//AuthorizerOptionRequestModel - -----------------------
type AuthorizerOptionRequestModel struct {
	RequestUserInfoRequestModel
	OptionName string `json:"option_name"`
}
//AuthorizerOptionReponseModel is
type AuthorizerOptionReponseModel struct {
	AuthorizerAppid string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
}
// AuthorizerOptionSETRequestModel is
type AuthorizerOptionSETRequestModel struct {
	AuthorizerOptionRequestModel
	OptionValue string `json:"option_value"`
}
//WXResponse is
type WXResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
