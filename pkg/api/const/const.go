package api

import (
	"fmt"

	"github.com/fatedier/frp/pkg/util/version"
)

var Endpoints = []string{
	"https://frp.api.locyanfrp.cn/v1/",
}

var UserAgents = struct {
	Client string
	Server string
}{
	Client: fmt.Sprintf("LoCyanFrp/%s (Frp Client; %s)", version.ApiService(), version.Full()),
	Server: fmt.Sprintf("LoCyanFrp/%s (Frp Server; %s)", version.ApiService(), version.Full()),
}
