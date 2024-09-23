package api

import (
	"crypto/tls"
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
	response := ResGetProxyCfg{}
	if err = json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	if response.Status != 200 {
		return "", ErrCheckTokenFail{response.Message}
	}
	return response.Data.Config, nil
}

// CheckFrpToken 校验客户端 Frp Token
func (s Service) CheckFrpToken(frpToken string, stk string) (ok bool, err error) {
	values := url.Values{}
	values.Set("frp_token", frpToken)
	values.Set("api_token", stk)
	s.Host.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(&s.Host)

	tr := &http.Transport{
		// 跳过证书验证
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(s.Host.String())
	// 请求出现错误，resp返回nil判断
	if resp == nil {
		return false, err
	}

	defer resp.Body.Close()
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, ErrHTTPStatus{
			Status: resp.StatusCode,
			Text:   resp.Status,
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	response := ResCheckFrpToken{}
	if err = json.Unmarshal(body, &response); err != nil {
		return false, err
	}
	if response.Status != 200 {
		return false, ErrCheckTokenFail{response.Message}
	}
	return true, nil
}
