package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

// V2Service LoCyanFrp API v2 service
type V2Service struct {
	Host url.URL
}

// ProxyStartGetCfg 简单启动获取Cfg
func (s Service) ProxyStartGetCfg(token string, proxyId string) (cfg string, err error) {
	values := url.Values{}
	values.Set("frp_token", token)
	values.Set("proxy_id", proxyId)
	// Encode 请求参数
	s.Host.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(&s.Host)

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(s.Host.String())
	// 请求出现错误，resp返回nil判断
	if resp == nil {
		return "", err
	}

	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", ErrHTTPStatus{
			Status: resp.StatusCode,
			Text:   resp.Status,
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	response := ResProxyCfg{}
	if err = json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	if response.Status != 200 {
		return "", ErrCheckTokenFail{response.Message}
	}
	return response.Data.Config, nil
}

type ResProxyCfg struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Config string `json:"config"`
	}
}
