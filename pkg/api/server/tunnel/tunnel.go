package tunnel

type PostTunnelParams struct {
	NodeId     int64
	FrpToken   string
	TunnelName string
	TunnelType string
	RemotePort *int
	Domain     *[]string
	Locations  *[]string
	SecretKey  *string
}

type PostTunnelResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		TunnelId int64 `json:"tunnel_id"`
	} `json:"data"`
}

func PostTunnel(apiKey string, params PostTunnelParams) (response PostTunnelResponse, err error) {
	// TODO
}
