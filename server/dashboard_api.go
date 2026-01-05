// Copyright 2017 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"cmp"
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/fatedier/frp/pkg/database"
	"github.com/fatedier/frp/pkg/msg"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/fatedier/frp/pkg/config/types"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/metrics/mem"
	httppkg "github.com/fatedier/frp/pkg/util/http"
	"github.com/fatedier/frp/pkg/util/log"
	netpkg "github.com/fatedier/frp/pkg/util/net"
	"github.com/fatedier/frp/pkg/util/version"
)

type GeneralResponse struct {
	Code int
	Msg  string
}

func (svr *Service) registerRouteHandlers(helper *httppkg.RouterRegisterHelper) {
	helper.Router.HandleFunc("/healthz", svr.healthz)
	subRouter := helper.Router.NewRoute().Subrouter()

	subRouter.Use(helper.AuthMiddleware.Middleware)

	// metrics
	if svr.cfg.EnablePrometheus {
		subRouter.Handle("/metrics", promhttp.Handler())
	}

	// apis
	subRouter.HandleFunc("/api/serverinfo", svr.apiServerInfo).Methods("GET")
	subRouter.HandleFunc("/api/proxy/{type}", svr.apiProxyByType).Methods("GET")
	subRouter.HandleFunc("/api/proxy/{type}/{name}", svr.apiProxyByTypeAndName).Methods("GET")
	subRouter.HandleFunc("/api/traffic/{name}", svr.apiProxyTraffic).Methods("GET")
	subRouter.HandleFunc("/api/proxies", svr.deleteProxies).Methods("DELETE")
	subRouter.HandleFunc("/api/proxies/close/{runId}", svr.CloseProxy).Methods("GET")
	subRouter.HandleFunc("/api/blacklist/list", svr.ShowClosedProxy).Methods("GET")
	subRouter.HandleFunc("/api/blacklist/add/{proxy_name}", svr.AddProxyFromDatabase).Methods("GET")
	subRouter.HandleFunc("/api/blacklist/delete/{proxy_name}", svr.DeleteProxyFromDatabase).Methods("GET")
	// view
	subRouter.Handle("/favicon.ico", http.FileServer(helper.AssetsFS)).Methods("GET")
	subRouter.PathPrefix("/static/").Handler(
		netpkg.MakeHTTPGzipHandler(http.StripPrefix("/static/", http.FileServer(helper.AssetsFS))),
	).Methods("GET")

	subRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})
}

type serverInfoResp struct {
	Version               string `json:"version"`
	BindPort              int    `json:"bindPort"`
	VhostHTTPPort         int    `json:"vhostHTTPPort"`
	VhostHTTPSPort        int    `json:"vhostHTTPSPort"`
	TCPMuxHTTPConnectPort int    `json:"tcpmuxHTTPConnectPort"`
	KCPBindPort           int    `json:"kcpBindPort"`
	QUICBindPort          int    `json:"quicBindPort"`
	SubdomainHost         string `json:"subdomainHost"`
	MaxPoolCount          int64  `json:"maxPoolCount"`
	MaxPortsPerClient     int64  `json:"maxPortsPerClient"`
	HeartBeatTimeout      int64  `json:"heartbeatTimeout"`
	AllowPortsStr         string `json:"allowPortsStr,omitempty"`
	TLSForce              bool   `json:"tlsForce,omitempty"`

	TotalTrafficIn  int64            `json:"totalTrafficIn"`
	TotalTrafficOut int64            `json:"totalTrafficOut"`
	CurConns        int64            `json:"curConns"`
	ClientCounts    int64            `json:"clientCounts"`
	ProxyTypeCounts map[string]int64 `json:"proxyTypeCount"`
}

// /healthz
func (svr *Service) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
}

// /api/serverinfo
func (svr *Service) apiServerInfo(w http.ResponseWriter, r *http.Request) {
	res := GeneralResponse{Code: 200}
	defer func() {
		log.Infof("http response [%s]: code [%d]", r.URL.Path, res.Code)
		w.WriteHeader(res.Code)
		if len(res.Msg) > 0 {
			_, _ = w.Write([]byte(res.Msg))
		}
	}()

	log.Infof("http request: [%s]", r.URL.Path)
	serverStats := mem.StatsCollector.GetServer()
	svrResp := serverInfoResp{
		Version:               version.Full(),
		BindPort:              svr.cfg.BindPort,
		VhostHTTPPort:         svr.cfg.VhostHTTPPort,
		VhostHTTPSPort:        svr.cfg.VhostHTTPSPort,
		TCPMuxHTTPConnectPort: svr.cfg.TCPMuxHTTPConnectPort,
		KCPBindPort:           svr.cfg.KCPBindPort,
		QUICBindPort:          svr.cfg.QUICBindPort,
		SubdomainHost:         svr.cfg.SubDomainHost,
		MaxPoolCount:          svr.cfg.Transport.MaxPoolCount,
		MaxPortsPerClient:     svr.cfg.MaxPortsPerClient,
		HeartBeatTimeout:      svr.cfg.Transport.HeartbeatTimeout,
		AllowPortsStr:         types.PortsRangeSlice(svr.cfg.AllowPorts).String(),
		TLSForce:              svr.cfg.Transport.TLS.Force,

		TotalTrafficIn:  serverStats.TotalTrafficIn,
		TotalTrafficOut: serverStats.TotalTrafficOut,
		CurConns:        serverStats.CurConns,
		ClientCounts:    serverStats.ClientCounts,
		ProxyTypeCounts: serverStats.ProxyTypeCounts,
	}

	buf, _ := json.Marshal(&svrResp)
	res.Msg = string(buf)
}

type BaseOutConf struct {
	v1.ProxyBaseConfig
}

type TCPOutConf struct {
	BaseOutConf
	RemotePort int `json:"remotePort"`
}

type TCPMuxOutConf struct {
	BaseOutConf
	v1.DomainConfig
	Multiplexer     string `json:"multiplexer"`
	RouteByHTTPUser string `json:"routeByHTTPUser"`
}

type UDPOutConf struct {
	BaseOutConf
	RemotePort int `json:"remotePort"`
}

type HTTPOutConf struct {
	BaseOutConf
	v1.DomainConfig
	Locations         []string `json:"locations"`
	HostHeaderRewrite string   `json:"hostHeaderRewrite"`
}

type HTTPSOutConf struct {
	BaseOutConf
	v1.DomainConfig
}

type STCPOutConf struct {
	BaseOutConf
}

type XTCPOutConf struct {
	BaseOutConf
}

func getConfByType(proxyType string) any {
	switch v1.ProxyType(proxyType) {
	case v1.ProxyTypeTCP:
		return &TCPOutConf{}
	case v1.ProxyTypeTCPMUX:
		return &TCPMuxOutConf{}
	case v1.ProxyTypeUDP:
		return &UDPOutConf{}
	case v1.ProxyTypeHTTP:
		return &HTTPOutConf{}
	case v1.ProxyTypeHTTPS:
		return &HTTPSOutConf{}
	case v1.ProxyTypeSTCP:
		return &STCPOutConf{}
	case v1.ProxyTypeXTCP:
		return &XTCPOutConf{}
	default:
		return nil
	}
}

// ProxyStatsInfo Get proxy info.
type ProxyStatsInfo struct {
	Name            string `json:"name"`
	Conf            any    `json:"conf"`
	ClientVersion   string `json:"clientVersion,omitempty"`
	TodayTrafficIn  int64  `json:"todayTrafficIn"`
	TodayTrafficOut int64  `json:"todayTrafficOut"`
	CurConns        int64  `json:"curConns"`
	LastStartTime   string `json:"lastStartTime"`
	LastCloseTime   string `json:"lastCloseTime"`
	Status          string `json:"status"`
}

type GetProxyInfoResp struct {
	Proxies []*ProxyStatsInfo `json:"proxies"`
}

// /api/proxy/:type
func (svr *Service) apiProxyByType(w http.ResponseWriter, r *http.Request) {
	res := GeneralResponse{Code: 200}
	params := mux.Vars(r)
	proxyType := params["type"]

	defer func() {
		log.Infof("http response [%s]: code [%d]", r.URL.Path, res.Code)
		w.WriteHeader(res.Code)
		if len(res.Msg) > 0 {
			_, _ = w.Write([]byte(res.Msg))
		}
	}()
	log.Infof("http request: [%s]", r.URL.Path)

	proxyInfoResp := GetProxyInfoResp{}
	proxyInfoResp.Proxies = svr.getProxyStatsByType(proxyType)
	slices.SortFunc(proxyInfoResp.Proxies, func(a, b *ProxyStatsInfo) int {
		return cmp.Compare(a.Name, b.Name)
	})

	buf, _ := json.Marshal(&proxyInfoResp)
	res.Msg = string(buf)
}

func (svr *Service) getProxyStatsByType(proxyType string) (proxyInfos []*ProxyStatsInfo) {
	proxyStats := mem.StatsCollector.GetProxiesByType(proxyType)
	proxyInfos = make([]*ProxyStatsInfo, 0, len(proxyStats))
	for _, ps := range proxyStats {
		proxyInfo := &ProxyStatsInfo{}
		if pxy, ok := svr.pxyManager.GetByName(ps.Name); ok {
			content, err := json.Marshal(pxy.GetConfigurer())
			if err != nil {
				log.Warnf("marshal proxy [%s] conf info error: %v", ps.Name, err)
				continue
			}
			proxyInfo.Conf = getConfByType(ps.Type)
			if err = json.Unmarshal(content, &proxyInfo.Conf); err != nil {
				log.Warnf("unmarshal proxy [%s] conf info error: %v", ps.Name, err)
				continue
			}
			proxyInfo.Status = "online"
			if pxy.GetLoginMsg() != nil {
				proxyInfo.ClientVersion = pxy.GetLoginMsg().Version
			}
		} else {
			proxyInfo.Status = "offline"
		}
		proxyInfo.Name = ps.Name
		proxyInfo.TodayTrafficIn = ps.TodayTrafficIn
		proxyInfo.TodayTrafficOut = ps.TodayTrafficOut
		proxyInfo.CurConns = ps.CurConns
		proxyInfo.LastStartTime = ps.LastStartTime
		proxyInfo.LastCloseTime = ps.LastCloseTime
		proxyInfos = append(proxyInfos, proxyInfo)
	}
	return
}

// GetProxyStatsResp Get proxy info by name.
type GetProxyStatsResp struct {
	Name            string `json:"name"`
	Conf            any    `json:"conf"`
	TodayTrafficIn  int64  `json:"todayTrafficIn"`
	TodayTrafficOut int64  `json:"todayTrafficOut"`
	CurConns        int64  `json:"curConns"`
	LastStartTime   string `json:"lastStartTime"`
	LastCloseTime   string `json:"lastCloseTime"`
	Status          string `json:"status"`
}

// /api/proxy/:type/:name
func (svr *Service) apiProxyByTypeAndName(w http.ResponseWriter, r *http.Request) {
	res := GeneralResponse{Code: 200}
	params := mux.Vars(r)
	proxyType := params["type"]
	name := params["name"]

	defer func() {
		log.Infof("http response [%s]: code [%d]", r.URL.Path, res.Code)
		w.WriteHeader(res.Code)
		if len(res.Msg) > 0 {
			_, _ = w.Write([]byte(res.Msg))
		}
	}()
	log.Infof("http request: [%s]", r.URL.Path)

	var proxyStatsResp GetProxyStatsResp
	proxyStatsResp, res.Code, res.Msg = svr.getProxyStatsByTypeAndName(proxyType, name)
	if res.Code != 200 {
		return
	}

	buf, _ := json.Marshal(&proxyStatsResp)
	res.Msg = string(buf)
}

func (svr *Service) getProxyStatsByTypeAndName(proxyType string, proxyName string) (proxyInfo GetProxyStatsResp, code int, msg string) {
	proxyInfo.Name = proxyName
	ps := mem.StatsCollector.GetProxiesByTypeAndName(proxyType, proxyName)
	if ps == nil {
		code = 404
		msg = "no proxy info found"
	} else {
		if pxy, ok := svr.pxyManager.GetByName(proxyName); ok {
			content, err := json.Marshal(pxy.GetConfigurer())
			if err != nil {
				log.Warnf("marshal proxy [%s] conf info error: %v", ps.Name, err)
				code = 400
				msg = "parse conf error"
				return
			}
			proxyInfo.Conf = getConfByType(ps.Type)
			if err = json.Unmarshal(content, &proxyInfo.Conf); err != nil {
				log.Warnf("unmarshal proxy [%s] conf info error: %v", ps.Name, err)
				code = 400
				msg = "parse conf error"
				return
			}
			proxyInfo.Status = "online"
		} else {
			proxyInfo.Status = "offline"
		}
		proxyInfo.TodayTrafficIn = ps.TodayTrafficIn
		proxyInfo.TodayTrafficOut = ps.TodayTrafficOut
		proxyInfo.CurConns = ps.CurConns
		proxyInfo.LastStartTime = ps.LastStartTime
		proxyInfo.LastCloseTime = ps.LastCloseTime
		code = 200
	}

	return
}

// GetProxyTrafficResp /api/traffic/:name
type GetProxyTrafficResp struct {
	Name       string  `json:"name"`
	TrafficIn  []int64 `json:"trafficIn"`
	TrafficOut []int64 `json:"trafficOut"`
}

func (svr *Service) apiProxyTraffic(w http.ResponseWriter, r *http.Request) {
	res := GeneralResponse{Code: 200}
	params := mux.Vars(r)
	name := params["name"]

	defer func() {
		log.Infof("http response [%s]: code [%d]", r.URL.Path, res.Code)
		w.WriteHeader(res.Code)
		if len(res.Msg) > 0 {
			_, _ = w.Write([]byte(res.Msg))
		}
	}()
	log.Infof("http request: [%s]", r.URL.Path)

	trafficResp := GetProxyTrafficResp{}
	trafficResp.Name = name
	proxyTrafficInfo := mem.StatsCollector.GetProxyTraffic(name)

	if proxyTrafficInfo == nil {
		res.Code = 404
		res.Msg = "no proxy info found"
		return
	}
	trafficResp.TrafficIn = proxyTrafficInfo.TrafficIn
	trafficResp.TrafficOut = proxyTrafficInfo.TrafficOut

	buf, _ := json.Marshal(&trafficResp)
	res.Msg = string(buf)
}

// DELETE /api/proxies?status=offline
func (svr *Service) deleteProxies(w http.ResponseWriter, r *http.Request) {
	res := GeneralResponse{Code: 200}

	log.Infof("http request: [%s]", r.URL.Path)
	defer func() {
		log.Infof("http response [%s]: code [%d]", r.URL.Path, res.Code)
		w.WriteHeader(res.Code)
		if len(res.Msg) > 0 {
			_, _ = w.Write([]byte(res.Msg))
		}
	}()

	status := r.URL.Query().Get("status")
	if status != "offline" {
		res.Code = 400
		res.Msg = "status only support offline"
		return
	}
	cleared, total := mem.StatsCollector.ClearOfflineProxies()
	log.Infof("cleared [%d] offline proxies, total [%d] proxies", cleared, total)
}

// CloseProxy GET /api/proxies/close/{runId}
func (svr *Service) CloseProxy(w http.ResponseWriter, r *http.Request) {

	// CloseProxy 数据库维护
	manager, err := database.NewClosedProxyManager("./closed_proxies.db")
	if err != nil {
		log.Infof(err.Error())
	}
	defer manager.Close()

	res := GeneralResponse{Code: 200}
	log.Debugf("http request: [%s]", r.URL.Path)

	defer func() {
		log.Infof("http response [%s]: code [%d]", r.URL.Path, res.Code)
		w.WriteHeader(res.Code)
		if len(res.Msg) > 0 {
			_, _ = w.Write([]byte(res.Msg))
		}
	}()

	params := mux.Vars(r)
	runId := params["runId"]
	if runId == "" {
		res.Code = 400
		res.Msg = "Please provide a valid run id"
		return
	}
	userType := r.URL.Query().Get("type")
	if userType == "" {
		res.Code = 400
		res.Msg = "Please provide a user type"
		return
	} else if userType != "admin" && userType != "user" {
		res.Code = 400
		res.Msg = "Please provide a valid user type"
	}

	// 这里仅能判断 runId
	isClosed, err := manager.IsClosed(runId)
	if err != nil {
		res.Code = 400
		res.Msg = "Can't search runId in database: " + err.Error()
		return
	}
	if isClosed {
		res.Code = 200
		res.Msg = "Already closed"
		return
	}

	// Get Proxy Control
	ctl, ok := svr.ctlManager.GetByID(runId)
	if !ok {
		// 无法通过 runId 查找就按照proxy name查找
		if pxy, ok := svr.pxyManager.GetByName(runId); ok {
			ctl, ok = svr.ctlManager.GetByID(pxy.GetLoginMsg().RunID)
			if !ok {
				res.Code = 400
				res.Msg = "Can't find proxy by proxy name: " + runId
				return
			}
		} else {
			res.Code = 400
			res.Msg = "Can't find proxy by proxy name: " + runId
			return
		}
	}

	// 非官方客户端需要加黑，若 tag 为 admin 则加黑, user 则不加黑允许重连
	if userType == "admin" {
		for range ctl.proxies {
			err := manager.AddClosedProxy(runId)
			if err != nil {
				res.Code = 400
				res.Msg = "Can’t add closed proxy to database: " + err.Error()
				return
			}
		}
	}

	// 判断是否为官方客户端, 如果为官方客户端可以让客户端强制关闭
	// 官方客户端因为不会自动重连，所以不执行加黑操作
	if strings.Contains(ctl.loginMsg.Version, "LoCyanFrp") {
		CloseInfo := &msg.CloseClient{RunId: runId}
		if err := ctl.msgDispatcher.Send(CloseInfo); err != nil {
			res.Code = 400
			res.Msg = "Can’t closed proxy: " + err.Error()
			return
		}
	} else {
		err3 := ctl.Close()
		if err3 != nil {
			res.Code = 400
			res.Msg = "Can’t close proxy: " + err3.Error()
			return
		}
	}

	res.Code = 200
	res.Msg = "OK"
	return
}

// DeleteProxyFromDatabase GET /api/proxies/delete/{runId}
func (svr *Service) DeleteProxyFromDatabase(w http.ResponseWriter, r *http.Request) {

	// CloseProxy 数据库维护
	manager, err := database.NewClosedProxyManager("./closed_proxies.db")
	if err != nil {
		log.Errorf("Failed to open database: %v", err)
		// 如果数据库打不开，应该直接返回 500，防止后续空指针
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer manager.Close()

	res := GeneralResponse{Code: 200}
	log.Debugf("http request: [%s]", r.URL.Path)

	defer func() {
		log.Infof("http response [%s]: code [%d]", r.URL.Path, res.Code)
		w.WriteHeader(res.Code)
		if len(res.Msg) > 0 {
			_, _ = w.Write([]byte(res.Msg))
		}
	}()

	params := mux.Vars(r)
	runId := params["runId"]
	if runId == "" {
		res.Code = 400
		res.Msg = "Please provide a valid runId"
		return
	}

	// 先检查是否存在 (适配原有逻辑: 如果找不到则返回错误)
	exists, err := manager.IsClosed(runId)
	if err != nil {
		res.Code = 500
		res.Msg = "Database error: " + err.Error()
		return
	}
	if !exists {
		res.Code = 404 // 或者 404
		res.Msg = "Can't find proxy with this runId"
		return
	}

	// 执行删除
	err = manager.DeleteClosed(runId)
	if err != nil {
		res.Code = 500
		res.Msg = "Can't delete proxy: " + err.Error()
		return
	}

	res.Code = 200
	res.Msg = "OK"
	return
}

func (svr *Service) ShowClosedProxy(w http.ResponseWriter, r *http.Request) {
	res := GeneralResponse{Code: 200}

	log.Infof("http request: [%s]", r.URL.Path)
	defer func() {
		log.Infof("http response [%s]: code [%d]", r.URL.Path, res.Code)
		w.WriteHeader(res.Code)
		if len(res.Msg) > 0 {
			_, _ = w.Write([]byte(res.Msg))
		}
	}()

	// CloseProxy 数据库维护
	manager, err := database.NewClosedProxyManager("./closed_proxies.db")
	if err != nil {
		log.Infof(err.Error())
	}
	defer manager.Close()

	proxies, err := manager.GetAllClosedProxies()
	if err != nil {
		res.Code = 500
		res.Msg = "Can't get proxies: " + err.Error()
		return
	}

	jsonData, err := json.Marshal(proxies)
	if err != nil {
		res.Code = 500
		res.Msg = "Can't get proxies: " + err.Error()
		return
	}

	res.Code = 200
	res.Msg = string(jsonData)
	return
}

// AddProxyFromDatabase GET /api/proxies/delete/{runId}
func (svr *Service) AddProxyFromDatabase(w http.ResponseWriter, r *http.Request) {

	// CloseProxy 数据库维护
	manager, err := database.NewClosedProxyManager("./closed_proxies.db")
	if err != nil {
		log.Infof(err.Error())
	}
	defer manager.Close()

	res := GeneralResponse{Code: 200}
	log.Debugf("http request: [%s]", r.URL.Path)

	defer func() {
		log.Infof("http response [%s]: code [%d]", r.URL.Path, res.Code)
		w.WriteHeader(res.Code)
		if len(res.Msg) > 0 {
			_, _ = w.Write([]byte(res.Msg))
		}
	}()

	// 必要参数
	params := mux.Vars(r)
	runId := params["runId"]
	if runId == "" {
		res.Code = 400
		res.Msg = "Please provide a valid runId"
		return
	}
	userType := r.URL.Query().Get("type")
	if userType == "" {
		res.Code = 400
		res.Msg = "Please provide a user type"
		return
	} else if userType != "admin" && userType != "user" {
		res.Code = 400
		res.Msg = "Please provide a valid user type"
	}

	// 先在本地检索是否存在该名称的代理
	_, ok := svr.ctlManager.GetByID(runId)
	if !ok {
		res.Code = 404
		res.Msg = "Proxy not found or not online, cannot retrieve RunID"
		return
	}

	if err := manager.AddClosedProxyWithType(runId, database.ClosedProxyType(userType)); err != nil {
		res.Code = 500
		res.Msg = "Database error: " + err.Error()
		return
	}

	res.Code = 200
	res.Msg = "OK"
	return
}
