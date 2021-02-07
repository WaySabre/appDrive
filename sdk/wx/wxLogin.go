package wxLogin

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/WaySabre/appDrive/sdk/unils"
)

const (
	UserInfoURL = "https://api.weixin.qq.com/sns/userinfo"
)

type (
	// WxUserInfo 微信用户资料
	WxUserInfo struct {
		OpenID     string `json:"openid,omitempty"`     // 授权用户唯一标识
		NickName   string `json:"nickname,omitempty"`   // 普通用户昵称
		Sex        uint32 `json:"sex,omitempty"`        // 普通用户性别，1为男性，2为女性
		Province   string `json:"province,omitempty"`   // 普通用户个人资料填写的省份
		City       string `json:"city,omitempty"`       // 普通用户个人资料填写的城市
		Country    string `json:"country,omitempty"`    // 国家，如中国为CN
		HeadImgURL string `json:"headimgurl,omitempty"` // 用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0代表640*640正方形头像），用户没有头像时该项为空
		//Privilege  string `json:"privilege"`
		Privilege []string `json:"privilege,omitempty"` // 用户特权信息，json数组，如微信沃卡用户为（chinaunicom）
		UnionID   string   `json:"unionid,omitempty"`   // 普通用户的标识，对当前开发者帐号唯一
		ErrCode   uint     `json:"errcode,omitempty"`
		ErrMsg    string   `json:"errmsg,omitempty"`
	}
)

// GetUserInfo 获取用户资料
func GetUserInfo(token string, openId string) (wxUserInfo *WxUserInfo, err error) {
	params := url.Values{
		"access_token": []string{token},
		"openid":       []string{openId},
	}
	body, err := utils.NewRequest("GET", UserInfoURL, []byte(params.Encode()))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &wxUserInfo)
	if err != nil {
		return nil, err
	}
	if wxUserInfo.OpenID == "" {
		return wxUserInfo, errors.New(wxUserInfo.ErrMsg)
	}
	return
}
