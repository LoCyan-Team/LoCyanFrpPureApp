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
	Client: fmt.Sprintf("LoCyanFrp-Client/3 (Frp Client; %s)", version.Full()),
	Server: fmt.Sprintf("LoCyanFrp-Server/3 (Frp Server; %s)", version.Full()),
}
