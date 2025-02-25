package api

import (
	"fmt"
	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/version"
	"io"
	"net/http"
	"net/url"
	"time"
)

func BoolToString(val bool) (str string) {
	if val {
		return "true"
	}
	return "false"
}

var apiV2Url = "https://api.locyanfrp.cn/v2/frp"
var apiV2BackupUrl = "https://backup.api.locyanfrp.cn/v2/frp"
var tr = &http.Transport{
	DisableKeepAlives: true,
}
var timeout = time.Second * 5
var ua = fmt.Sprintf("LoCyanFrp/1.0 (Frp; %s)", version.FullText())

func createRequest(method string, api url.URL, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, api.String(), body)
	if err != nil {
		return nil, err
	}
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("User-Agent", ua)
	return req, err
}

func RequestApi(endpoint string, method string, queryParams url.Values, body io.Reader) (*http.Response, error) {
	client := &http.Client{Transport: tr, Timeout: timeout}

	// 请求主 API
	api, _ := url.Parse(apiV2Url + endpoint)

	api.RawQuery = queryParams.Encode()
	defer func(u *url.URL) {
		u.RawQuery = ""
	}(api)

	req, err := createRequest(method, *api, body)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Warn("request API failed, switching to backup API")
		// 请求备用 API
		backupApi, _ := url.Parse(apiV2BackupUrl + endpoint)

		backupApi.RawQuery = queryParams.Encode()
		defer func(u *url.URL) {
			u.RawQuery = ""
		}(backupApi)

		backupReq, err := createRequest(method, *backupApi, body)
		if err != nil {
			return nil, err
		}

		backupRes, err := client.Do(backupReq)
		if err != nil {
			return nil, err
		}
		return backupRes, nil
	}
	return res, nil
}
