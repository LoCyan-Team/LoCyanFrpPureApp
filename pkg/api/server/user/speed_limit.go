package user

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

func (s Service) GetSpeedLimit(apiKey string, params GetSpeedLimitParams) (response GetSpeedLimitResponse, err error) {
	// TODO
}
