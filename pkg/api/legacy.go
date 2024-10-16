package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/fatedier/frp/pkg/msg"
)

// Service LoCyanFrp Frp Server API service
type Service struct {
	Host url.URL
}

// NewService LoCyanFrp Frp Server API service
func NewService(host string) (s *Service, err error) {
	u, err := url.Parse(host)
	if err != nil {
		return
	}
	return &Service{*u}, nil
}

// EZStartGetCfg 简单启动获取Cfg
// 已废弃，请改用 ProxyStartGetCfg
// Deprecated
func (s Service) EZStartGetCfg(token string, proxyId string) (cfg string, err error) {
	values := url.Values{}
	values.Set("action", "getcfg")
	values.Set("token", token)
	values.Set("id", proxyId)
	// Encode 请求参数
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
	response := ResGetCfg{}
	if err = json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	if !response.Success {
		return "", ErrCheckTokenFail{response.Message}
	}
	return response.Cfg, nil
}

// SubmitRunId 提交runID至服务器
func (s Service) SubmitRunId(stk string, pMsg *msg.NewProxy, runId string) (err error) {
	values := url.Values{}
	values.Set("run_id", runId)
	// user frpToken
	values.Set("proxy_name", pMsg.ProxyName)
	values.Set("apitoken", stk)
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
		return err
	}

	// 提交就完事了管他那么多干什么
	defer resp.Body.Close()
	return err
}

// CheckToken 校验客户端 token
func (s Service) CheckToken(user string, token string, timestamp int64, stk string) (ok bool, err error) {
	values := url.Values{}
	values.Set("action", "checktoken")
	values.Set("user", user)
	values.Set("token", token)
	values.Set("timestamp", fmt.Sprintf("%d", timestamp))
	values.Set("apitoken", stk)
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
	response := ResponseCheckToken{}
	if err = json.Unmarshal(body, &response); err != nil {
		return false, err
	}
	if !response.Success {
		return false, ErrCheckTokenFail{response.Message}
	}
	return true, nil
}

// CheckProxy 校验客户端代理
func (s Service) CheckProxy(user string, pMsg *msg.NewProxy, timestamp int64, stk string) (ok bool, err error) {
	domains, err := json.Marshal(pMsg.CustomDomains)
	if err != nil {
		return false, err
	}

	headers, err := json.Marshal(pMsg.Headers)
	if err != nil {
		return false, err
	}

	locations, err := json.Marshal(pMsg.Locations)
	if err != nil {
		return false, err
	}

	values := url.Values{}

	// API Basic
	values.Set("action", "checkproxy")
	values.Set("user", user)
	values.Set("timestamp", fmt.Sprintf("%d", timestamp))
	values.Set("apitoken", stk)

	// Proxies basic info
	values.Set("proxy_name", pMsg.ProxyName)
	values.Set("proxy_type", pMsg.ProxyType)
	values.Set("use_encryption", BoolToString(pMsg.UseEncryption))
	values.Set("use_compression", BoolToString(pMsg.UseCompression))

	// Http Proxies
	values.Set("domain", string(domains))
	//values.Set("subdomain", pMsg.SubDomain)

	// Headers
	values.Set("locations", string(locations))
	values.Set("http_user", pMsg.HTTPUser)
	values.Set("http_pwd", pMsg.HTTPPwd)
	values.Set("host_header_rewrite", pMsg.HostHeaderRewrite)
	values.Set("headers", string(headers))

	// Tcp & Udp & Stcp
	values.Set("remote_port", strconv.Itoa(pMsg.RemotePort))

	// Stcp & Xtcp
	values.Set("sk", pMsg.Sk)

	// Load balance
	values.Set("group", pMsg.Group)
	values.Set("group_key", pMsg.GroupKey)

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
	response := ResponseCheckProxy{}
	if err = json.Unmarshal(body, &response); err != nil {
		return false, err
	}
	if !response.Success {
		return false, ErrCheckProxyFail{response.Message}
	}
	return true, nil
}

// GetProxyLimit 获取隧道限速信息
func (s Service) GetProxyLimit(user string, timestamp int64, stk string) (inLimit, outLimit uint64, err error) {
	// 这部分就照之前的搬过去了，能跑就行x
	values := url.Values{}
	values.Set("action", "getlimit")
	values.Set("user", user)
	values.Set("timestamp", fmt.Sprintf("%d", timestamp))
	values.Set("apitoken", stk)
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

	defer resp.Body.Close()

	// 请求出现错误，resp返回nil判断
	if resp == nil {
		return 1280, 1280, err
	}

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

	response := &ResponseGetLimit{}
	if err = json.Unmarshal(body, response); err != nil {
		return 1280, 1280, err
	}

	// 这里直接返回 uint64 应该问题不大
	return response.MaxIn, response.MaxOut, nil
}

func (e ErrCheckTokenFail) Error() string {
	return e.Message
}

func (e ErrCheckProxyFail) Error() string {
	return e.Message
}
