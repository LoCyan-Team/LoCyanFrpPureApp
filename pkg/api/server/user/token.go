package user

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	_api "github.com/fatedier/frp/pkg/api/exec"
	"github.com/fatedier/frp/pkg/api/type"
)

type PostTokenParams struct {
	NodeId   int64  `json:"node_id"`
	FrpToken string `json:"frp_token"`
}

type PostTokenResponse struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Data    struct{} `json:"data"`
}

func (s Service) PostToken(serverConfig api.ServerConfig, params PostTokenParams) (response *PostTokenResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 设置超时
	defer cancel()

	body := &url.Values{}
	body.Set("node_id", strconv.FormatInt(params.NodeId, 10))
	body.Set("frp_token", params.FrpToken)
	rs, err := _api.ServerExecute(ctx, "/server/user/token", http.MethodPost, serverConfig, nil, body)
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
