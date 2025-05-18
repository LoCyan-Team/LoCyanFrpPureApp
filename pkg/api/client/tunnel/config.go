package tunnel

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

func (s Service) GetConfig(params GetConfigParams) (response GetConfigResponse, err error) {
	// TODO
}
