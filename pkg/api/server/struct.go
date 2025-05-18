package server

import (
	"github.com/fatedier/frp/pkg/api/server/tunnel"
	"github.com/fatedier/frp/pkg/api/server/user"
)

type Service struct {
	Tunnel tunnel.Service
	User   user.Service
}
