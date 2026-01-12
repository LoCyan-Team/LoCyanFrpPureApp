package tunnel

import (
	"context"
	"encoding/json"

	"io"
	"net/http"
	"time"

	_api "github.com/fatedier/frp/pkg/api/exec"
)

type GetConfigParams struct {
	FrpToken string `json:"frp_token"`
	TunnelId int64  `json:"tunnel_id"`
}

type GetConfigResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Config string `json:"config"`
	} `json:"data"`
}

func (s Service) GetConfig(params GetConfigParams) (response *GetConfigResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 设置超时
	defer cancel()

	rs, err := _api.ClientExecute(ctx, "/client/tunnel/config", http.MethodGet, params, nil)
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
