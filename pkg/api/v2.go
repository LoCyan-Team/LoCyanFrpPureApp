package api

import (
	"encoding/json"
	"fmt"
	"github.com/fatedier/frp/pkg/msg"
	"github.com/fatedier/frp/pkg/util/version"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// V2Service LoCyanFrp Frp API v2
type V2Service struct {
}

var apiV2Url = "https://api.locyanfrp.cn/v2/frp"
var tr = &http.Transport{
	DisableKeepAlives: true,
}
var ua = fmt.Sprintf("LoCyanFrp/1.0 (Frp; %s)", version.FullText())

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

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(http.MethodGet, api.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 非正常状态码
	if resp.StatusCode != http.StatusOK {
		errInfo := ResError{}
		if err = json.Unmarshal(body, &errInfo); err != nil {
			return "", err
		}
		return "", errInfo
	}

	response := ResGetProxyCfg{}
	if err = json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	return response.Data.Config, nil
}

// SubmitRunId 提交runID至服务器
func (s V2Service) SubmitRunId(apiToken string, nodeId int, pMsg *msg.NewProxy, runId string) (err error) {
	api, _ := url.Parse(apiV2Url + "/server/run-id")
	values := url.Values{}

	name := strings.Split(pMsg.ProxyName, ".")[1]

	values.Set("run_id", runId)
	values.Set("proxy_name", name)
	values.Set("api_token", apiToken+"|"+strconv.Itoa(nodeId))

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(http.MethodPost, api.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// FrpTokenCheck 校验客户端 Frp Token
func (s V2Service) FrpTokenCheck(frpToken string, apiToken string, nodeId int) (ok bool, err error) {
	api, _ := url.Parse(apiV2Url + "/server/token")
	values := url.Values{}
	values.Set("frp_token", frpToken)
	values.Set("api_token", apiToken+"|"+strconv.Itoa(nodeId))
	api.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(api)

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(http.MethodGet, api.String(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// 非正常状态码
	if resp.StatusCode != http.StatusOK {
		errInfo := ResError{}
		if err = json.Unmarshal(body, &errInfo); err != nil {
			return false, err
		}
		return false, errInfo
	}

	response := ResCheckFrpToken{}
	if err = json.Unmarshal(body, &response); err != nil {
		return false, err
	}
	return true, nil
}

// ProxyCheck 校验客户端代理
func (s V2Service) ProxyCheck(frpToken string, pMsg *msg.NewProxy, apiToken string, nodeId int) (ok bool, err error) {
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

	name := strings.Split(pMsg.ProxyName, ".")[1]

	// API Basic
	values.Set("frp_token", frpToken)
	values.Set("api_token", apiToken+"|"+strconv.Itoa(nodeId))

	// Proxies basic info
	values.Set("proxy_name", name)
	//log.Info("Proxy name: " + pMsg.ProxyName)
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
	//log.Info(strconv.Itoa(pMsg.RemotePort))

	// STCP & XTCP
	values.Set("secret_key", pMsg.Sk)

	// Load balance
	//values.Set("group", pMsg.Group)
	//values.Set("group_key", pMsg.GroupKey)

	api.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(api)

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(http.MethodGet, api.String(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// 非正常状态码
	if resp.StatusCode != http.StatusOK {
		errInfo := ResError{}
		if err = json.Unmarshal(body, &errInfo); err != nil {
			return false, err
		}
		return false, errInfo
	}

	response := ResCheckProxy{}
	if err = json.Unmarshal(body, &response); err != nil {
		return false, err
	}
	return true, nil
}

// GetLimit 获取隧道限速信息
func (s V2Service) GetLimit(frpToken string, apiToken string, nodeId int) (inLimit, outLimit uint64, err error) {
	api, _ := url.Parse(apiV2Url + "/server/limit")
	// 这部分就照之前的搬过去了，能跑就行x
	values := url.Values{}
	values.Set("frp_token", frpToken)
	values.Set("api_token", apiToken+"|"+strconv.Itoa(nodeId))
	api.RawQuery = values.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(api)

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(http.MethodGet, api.String(), nil)
	if err != nil {
		return 1280, 1280, err
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return 1280, 1280, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 1280, 1280, err
	}

	// 非正常状态码
	if resp.StatusCode != http.StatusOK {
		errInfo := ResError{}
		if err = json.Unmarshal(body, &errInfo); err != nil {
			return 1280, 1280, err
		}
		return 1280, 1280, errInfo
	}

	response := ResGetLimit{}
	if err = json.Unmarshal(body, &response); err != nil {
		return 1280, 1280, err
	}

	// 这里直接返回 uint64 应该问题不大
	return response.Data.Inbound, response.Data.Outbound, nil
}
