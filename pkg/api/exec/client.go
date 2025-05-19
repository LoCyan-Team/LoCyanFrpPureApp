package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/version"
	"io"
	"net/http"
	"net/url"
)

func Execute(ctx context.Context, path string, method string, apiKey string, params interface{}) (*http.Response, error) {
	var (
		fullURL string
		resp    *http.Response
		err     error
	)
	for _, baseURL := range V3ApiEndpoints {
		fullURL, err = url.JoinPath(baseURL, path)
		if err != nil {
			continue
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", fmt.Sprintf("LoCyanFrp/2.0 (Frp; %s)", version.Full()))
		if apiKey != "" {
			req.Header.Set("X-Node-API-Key", apiKey)
		}

		if params != nil {
			switch method {
			case http.MethodGet, http.MethodDelete:
				query := req.URL.Query()

				paramBytes, err := json.Marshal(params)
				if err != nil {
					return nil, err
				}

				var paramMap map[string]interface{}
				err = json.Unmarshal(paramBytes, &paramMap)
				if err != nil {
					return nil, err
				}

				for key, value := range paramMap {
					query.Add(key, fmt.Sprintf("%v", value))
				}

				req.URL.RawQuery = query.Encode()
			case http.MethodPost, http.MethodPut, http.MethodPatch:
				body, err := json.Marshal(params)
				if err != nil {
					return nil, err
				}
				req.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}

		resp, err = http.DefaultClient.Do(req)
		if err == nil {
			// 请求成功
			return resp, nil
		}

		log.Warnf("API Error: request failed (endpoint: %s)", baseURL)
		if resp != nil {
			resp.Body.Close()
		}
	}
	return nil, errors.New("API Error: request failed, all endpoints failed")
}
