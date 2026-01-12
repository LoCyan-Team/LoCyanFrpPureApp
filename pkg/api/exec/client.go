package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/fatedier/frp/pkg/api/const"
	"github.com/fatedier/frp/pkg/util/log"
)

func ClientExecute(ctx context.Context, path string, method string, query interface{}, body *url.Values) (*http.Response, error) {
	var (
		fullURL string
		resp    *http.Response
		err     error
	)
	for _, baseURL := range api.Endpoints {
		fullURL, err = url.JoinPath(baseURL, path)
		if err != nil {
			continue
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", api.UserAgents.Client)

		switch method {
		case http.MethodGet, http.MethodDelete:
			if query != nil {
				reqQuery := req.URL.Query()

				paramBytes, err := json.Marshal(query)
				if err != nil {
					return nil, err
				}

				var paramMap map[string]interface{}
				err = json.Unmarshal(paramBytes, &paramMap)
				if err != nil {
					return nil, err
				}

				for key, value := range paramMap {
					reqQuery.Add(key, fmt.Sprintf("%v", value))
				}

				req.URL.RawQuery = reqQuery.Encode()
			}
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			if body != nil {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Body = io.NopCloser(bytes.NewBufferString(body.Encode()))
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
