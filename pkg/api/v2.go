package api

import (
	"crypto/tls"
	"encoding/json"
	"github.com/fatedier/frp/pkg/msg"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// V2Service LoCyanFrp Frp API v2
type V2Service struct {
}

var apiV2Url = "https://api-v2.locyanfrp.cn/api/v2/frp"

// NewApiService LoCyanFrp API service
func NewApiService() (s *V2Service, err error) {
	return &V2Service{}, nil
}

// ProxyStartGetCfg 简单启动获取Cfg
func (s V2Service) ProxyStartGetCfg(frpToken string, proxyId string) (cfg string, err error) {
	api, _ := url.Parse(apiV2Url + "/client/config")
	values := url.Values{}
	values.Set("frp_token", frpToken)
	values.Set("proxy_id", proxyId)
	// Encode 请求参数
	api.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(api)

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(api.Host)
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

// SubmitRunId 提交runID至服务器
func (s V2Service) SubmitRunId(apiToken string, pMsg *msg.NewProxy, runId string) (err error) {
	api, _ := url.Parse(apiV2Url + "/server/run-id")
	values := url.Values{}
	values.Set("run_id", runId)
	values.Set("proxy_name", pMsg.ProxyName)
	values.Set("api_token", apiToken)
	api.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(api)

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Post(api.Host, "application/json", strings.NewReader(values.Encode()))
	// 请求出现错误，resp返回nil判断
	if resp == nil {
		return err
	}

	// 提交就完事了管他那么多干什么
	defer resp.Body.Close()
	return err
}

// CheckFrpToken 校验客户端 Frp Token
func (s V2Service) CheckFrpToken(frpToken string, apiToken string) (ok bool, err error) {
	api, _ := url.Parse(apiV2Url + "/server/token")
	values := url.Values{}
	values.Set("frp_token", frpToken)
	values.Set("api_token", apiToken)
	api.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(api)

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(api.Host)
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

// CheckProxy 校验客户端代理
func (s V2Service) CheckProxy(frpToken string, pMsg *msg.NewProxy, apiToken string) (ok bool, err error) {
	api, _ := url.Parse(apiV2Url + "/server/proxy")
	domains, err := json.Marshal(pMsg.CustomDomains)
	if err != nil {
		return false, err
	}

	//headers, err := json.Marshal(pMsg.Headers)
	//if err != nil {
	//	return false, err
	//}
	//
	//locations, err := json.Marshal(pMsg.Locations)
	//if err != nil {
	//	return false, err
	//}

	values := url.Values{}

	// API Basic
	values.Set("frp_token", frpToken)
	values.Set("api_token", apiToken)

	// Proxies basic info
	values.Set("proxy_name", pMsg.ProxyName)
	values.Set("proxy_type", pMsg.ProxyType)
	values.Set("use_encryption", BoolToString(pMsg.UseEncryption))
	values.Set("use_compression", BoolToString(pMsg.UseCompression))

	// Http Proxies
	values.Set("domain", string(domains))
	//values.Set("subdomain", pMsg.SubDomain)

	// Headers
	//values.Set("locations", string(locations))
	//values.Set("http_user", pMsg.HTTPUser)
	//values.Set("http_pwd", pMsg.HTTPPwd)
	//values.Set("host_header_rewrite", pMsg.HostHeaderRewrite)
	//values.Set("headers", string(headers))

	// TCP & UDP & STCP
	values.Set("remote_port", strconv.Itoa(pMsg.RemotePort))

	// STCP & XTCP
	values.Set("secret_key", pMsg.Sk)

	// Load balance
	//values.Set("group", pMsg.Group)
	//values.Set("group_key", pMsg.GroupKey)

	api.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(api)

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(api.Host)

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
	response := ResponseCheckProxy{}
	if err = json.Unmarshal(body, &response); err != nil {
		return false, err
	}
	if !response.Success {
		return false, ErrCheckProxyFail{response.Message}
	}
	return true, nil
}

// GetLimit 获取隧道限速信息
func (s V2Service) GetLimit(frpToken string, apiToken string) (inLimit, outLimit uint64, err error) {
	api, _ := url.Parse(apiV2Url + "/server/limit")
	// 这部分就照之前的搬过去了，能跑就行x
	values := url.Values{}
	values.Set("frp_token", frpToken)
	values.Set("api_token", apiToken)
	api.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(api)

	tr := &http.Transport{
		// 跳过证书验证
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(api.Host)

	defer resp.Body.Close()

	if err != nil {
		return 1280, 1280, err
	}
	if resp.StatusCode != http.StatusOK {
		return 1280, 1280, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 1280, 1280, err
	}

	er := &ErrHTTPStatus{}
	if err = json.Unmarshal(body, er); err != nil {
		return 1280, 1280, err
	}
	if er.Status != 200 {
		return 1280, 1280, er
	}

	response := &ResGetLimit{}
	if err = json.Unmarshal(body, response); err != nil {
		return 1280, 1280, err
	}

	// 这里直接返回 uint64 应该问题不大
	return response.Data.Inbound, response.Data.Outbound, nil
}
