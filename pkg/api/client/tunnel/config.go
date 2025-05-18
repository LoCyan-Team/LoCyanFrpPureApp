package tunnel

import (
	"context"
	"encoding/json"
	_api "github.com/fatedier/frp/pkg/api/exec"
	"io"
	"net/http"
	"time"
)

type GetConfigParams struct {
	FrpToken string
	TunnelId int64
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

	rs, err := _api.Execute(ctx, "/client/config", http.MethodPut, "", params)
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
