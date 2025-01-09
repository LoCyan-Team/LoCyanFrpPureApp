package api

import "fmt"

type ResGetProxyCfg struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Config string `json:"config"`
	}
}
type ResCheckFrpToken struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct{}
}

type ResCheckProxy struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct{}
}

type ResGetLimit struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Inbound  uint64 `json:"inbound"`
		Outbound uint64 `json:"outbound"`
	}
}

type ResError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (r ResError) Error() string {
	return fmt.Sprintf("LoCyanFrp API Error (Status: %d, Message: %s)", r.Status, r.Message)
}
