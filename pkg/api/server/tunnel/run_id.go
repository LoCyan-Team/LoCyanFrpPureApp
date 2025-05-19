package tunnel

import (
	"context"
	"encoding/json"
	_api "github.com/fatedier/frp/pkg/api/exec"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type PostRunIdParams struct {
	NodeId   int64  `json:"node_id"`
	TunnelId int64  `json:"tunnel_id"`
	RunId    string `json:"run_id"`
}

type PostRunIdResponse struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Data    struct{} `json:"data"`
}

func (s Service) PutRunId(apiKey string, params PostRunIdParams) (response *PostRunIdResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 设置超时
	defer cancel()

	body := &url.Values{}
	body.Set("node_id", strconv.FormatInt(params.NodeId, 10))
	body.Set("tunnel_id", strconv.FormatInt(params.TunnelId, 10))
	body.Set("run_id", params.RunId)
	rs, err := _api.Execute(ctx, "/server/tunnel/run-id", http.MethodPut, apiKey, nil, body)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	bodyBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, err
	}
	return response, nil
}
