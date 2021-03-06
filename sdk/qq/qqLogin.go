package qqLogin

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"github.com/WaySabre/appDrive/sdk/unils"
)

const (
	UserInfoURL = "http://openapi.tencentyun.com/v3/user/get_info"
)

type (
	// QQ 用户资料
	QqUserInfo struct {
		OpenID    string `json:"openid,omitempty"`    // 授权用户唯一标识
		NickName  string `json:"nickname,omitempty"`  // 普通用户昵称
		Gender    string `json:"gender,omitempty"`    // 普通用户性别
		Province  string `json:"province,omitempty"`  // 普通用户个人资料填写的省份
		City      string `json:"city,omitempty"`      // 普通用户个人资料填写的城市
		Country   string `json:"country,omitempty"`   // 国家，如中国为CN
		Figureurl string `json:"figureurl,omitempty"` // 用户头像
	}
)

type Config struct {
	Appid string
	Key   string
}

// GetUserInfo 获取用户资料
func (e *Config) GetUserInfo(openId, token, pf string) (QqUserInfo *QqUserInfo, err error) {
	params := url.Values{
		"appid":   []string{e.Appid},
		"format":  []string{"json"},
		"openid":  []string{openId},
		"openkey": []string{token},
		"pf":      []string{pf},
	}
	//获取源串
	sourceUrl := url.QueryEscape(params.Encode())
	sig := HmacSHA1(e.Key+"&", "GET&%2Fv3%2Fuser%2Fget_info&"+sourceUrl)
	params.Add("sig", sig)
	body, err := utils.NewRequest("GET", UserInfoURL, []byte(params.Encode()))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &QqUserInfo)
	if err != nil {
		return nil, err
	}
	QqUserInfo.Figureurl = strings.Replace(QqUserInfo.Figureurl, "\\", "", -1)
	QqUserInfo.OpenID = openId
	return
}

func HmacSHA1(keyStr, value string) string {
	key := []byte(keyStr)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(value))
	res := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return res
}
