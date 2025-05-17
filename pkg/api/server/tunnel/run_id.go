package tunnel

type PostRunIdParams struct {
	NodeId int64
	RunId  string
}

type PostRunIdResponse struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Data    struct{} `json:"data"`
}

func (s Service) PutRunId(apiKey string, params PostRunIdParams) (response PostRunIdResponse, err error) {
	// TODO
}
