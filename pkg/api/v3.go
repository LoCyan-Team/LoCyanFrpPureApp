package api

import (
	"github.com/fatedier/frp/pkg/api/client"
	"github.com/fatedier/frp/pkg/api/server"
)

type V3Service struct {
	Server server.Service
	Client client.Service
}

func NewApiService() (s *V3Service, err error) {
	return &V3Service{}, nil
}
