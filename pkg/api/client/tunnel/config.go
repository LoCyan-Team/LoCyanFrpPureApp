package tunnel

type GetConfigParams struct {
	FrpToken   string
	NodeId     int64
	TunnelName string
}

type GetConfigResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Config string `json:"config"`
	} `json:"data"`
}

func (s Service) GetConfig(params GetConfigParams) (response GetConfigResponse, err error) {
	// TODO
}
