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

type ResVerifyProxy struct {
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

// Legacy

type ErrHTTPStatus struct {
	Status int    `json:"status"`
	Text   string `json:"message"`
}

func (e ErrHTTPStatus) Error() string {
	return fmt.Sprintf("LoCyanFrp API Error (Status: %d, Text: %s)", e.Status, e.Text)
}

type ResponseGetLimit struct {
	MaxIn  uint64 `json:"max-in"`
	MaxOut uint64 `json:"max-out"`
}

type ResponseCheckToken struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ResponseCheckProxy struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ResGetCfg struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Cfg     string `json:"cfg"`
}

type ErrCheckTokenFail struct {
	Message string
}

type ErrCheckProxyFail struct {
	Message string
}
