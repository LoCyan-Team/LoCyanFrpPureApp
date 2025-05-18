package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/fatedier/frp/pkg/util/log"
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

		var reqBody []byte
		if params != nil {
			reqBody, err = json.Marshal(params)
			if err != nil {
				return nil, err
			}
		}

		req, _ := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(reqBody))

		req.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			req.Header.Set("X-Node-API-Key", apiKey)
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
