package tunnel

type PostTokenParams struct {
	NodeId   int64
	FrpToken string
}

type PostTokenResponse struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Data    struct{} `json:"data"`
}

func (s Service) PostToken(apiKey string, params PostTokenParams) (response PostTokenResponse, err error) {
}
