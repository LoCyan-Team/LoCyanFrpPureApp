package user

import (
	"context"
	"encoding/json"
	_api "github.com/fatedier/frp/pkg/api/exec"
	"io"
	"net/http"
	"time"
)

type GetSpeedLimitParams struct {
	NodeId   int64
	FrpToken string
}

type GetSpeedLimitResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Inbound  int64 `json:"inbound"`
		Outbound int64 `json:"outbound"`
	} `json:"data"`
}

func (s Service) GetSpeedLimit(apiKey string, params GetSpeedLimitParams) (response *GetSpeedLimitResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 设置超时
	defer cancel()

	rs, err := _api.Execute(ctx, "/server/speed-limit", http.MethodGet, apiKey, params)
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
