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

type PostTunnelParams struct {
	NodeId         int64    `json:"node_id"`
	FrpToken       string   `json:"frp_token"`
	TunnelName     string   `json:"tunnel_name"`
	TunnelType     string   `json:"tunnel_type"`
	RemotePort     *int     `json:"remote_port"`
	UseCompression bool     `json:"use_compression"`
	UseEncryption  bool     `json:"use_encryption"`
	Domain         []string `json:"domain"`
	Locations      []string `json:"locations"`
	SecretKey      *string  `json:"secret_key"`
}

type PostTunnelResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		TunnelId int64 `json:"tunnel_id"`
	} `json:"data"`
}

func (s Service) PostTunnel(apiKey string, params PostTunnelParams) (response *PostTunnelResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 设置超时
	defer cancel()

	body := &url.Values{}
	body.Set("node_id", strconv.FormatInt(params.NodeId, 10))
	body.Set("frp_token", params.FrpToken)
	body.Set("tunnel_name", params.TunnelName)
	body.Set("tunnel_type", params.TunnelType)
	body.Set("use_compression", strconv.FormatBool(params.UseCompression))
	body.Set("use_encryption", strconv.FormatBool(params.UseEncryption))
	if params.RemotePort != nil {
		body.Set("remote_port", strconv.Itoa(*params.RemotePort))
	}
	for _, domain := range params.Domain {
		body.Add("domain", domain)
	}
	for _, location := range params.Locations {
		body.Add("locations", location)
	}
	if params.SecretKey != nil {
		body.Set("secret_key", *params.SecretKey)
	}
	rs, err := _api.Execute(ctx, "/server/tunnel", http.MethodPost, apiKey, nil, body)
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
