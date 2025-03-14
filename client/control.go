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

package client

import (
	"context"
	"io"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/fatedier/golib/control/shutdown"
	"github.com/fatedier/golib/crypto"

	"github.com/fatedier/frp/client/proxy"
	"github.com/fatedier/frp/client/visitor"
	"github.com/fatedier/frp/pkg/auth"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/msg"
	"github.com/fatedier/frp/pkg/transport"
	"github.com/fatedier/frp/pkg/util/xlog"
)

type Control struct {
	// service context
	ctx context.Context
	xl  *xlog.Logger

	// Unique ID obtained from frps.
	// It should be attached to the login message when reconnecting.
	runID string

	// manage all proxies
	pxyCfgs map[string]config.ProxyConf
	pm      *proxy.Manager

	// manage all visitors
	vm *visitor.Manager

	// control connection
	conn net.Conn

	cm *ConnectionManager

	// put a message in this channel to send it over control connection to server
	sendCh chan (msg.Message)

	// read from this channel to get the next message sent by server
	readCh chan (msg.Message)

	// goroutines can block by reading from this channel, it will be closed only in reader() when control connection is closed
	closedCh chan struct{}

	closedDoneCh chan struct{}

	// last time got the Pong message
	lastPong time.Time

	// The client configuration
	clientCfg config.ClientCommonConf

	readerShutdown     *shutdown.Shutdown
	writerShutdown     *shutdown.Shutdown
	msgHandlerShutdown *shutdown.Shutdown

	// sets authentication based on selected method
	authSetter auth.Setter

	msgTransporter transport.MessageTransporter
}

func NewControl(
	ctx context.Context, runID string, conn net.Conn, cm *ConnectionManager,
	clientCfg config.ClientCommonConf,
	pxyCfgs map[string]config.ProxyConf,
	visitorCfgs map[string]config.VisitorConf,
	authSetter auth.Setter,
) *Control {
	// new xlog instance
	ctl := &Control{
		ctx:                ctx,
		xl:                 xlog.FromContextSafe(ctx),
		runID:              runID,
		conn:               conn,
		cm:                 cm,
		pxyCfgs:            pxyCfgs,
		sendCh:             make(chan msg.Message, 100),
		readCh:             make(chan msg.Message, 100),
		closedCh:           make(chan struct{}),
		closedDoneCh:       make(chan struct{}),
		clientCfg:          clientCfg,
		readerShutdown:     shutdown.New(),
		writerShutdown:     shutdown.New(),
		msgHandlerShutdown: shutdown.New(),
		authSetter:         authSetter,
	}
	ctl.msgTransporter = transport.NewMessageTransporter(ctl.sendCh)
	ctl.pm = proxy.NewManager(ctl.ctx, clientCfg, ctl.msgTransporter)

	ctl.vm = visitor.NewManager(ctl.ctx, ctl.runID, ctl.clientCfg, ctl.connectServer, ctl.msgTransporter)
	ctl.vm.Reload(visitorCfgs)
	return ctl
}

func (ctl *Control) Run() {
	go ctl.worker()

	// start all proxies
	ctl.pm.Reload(ctl.pxyCfgs)

	// start all visitors
	go ctl.vm.Run()
}

func (ctl *Control) HandleReqWorkConn(_ *msg.ReqWorkConn) {
	xl := ctl.xl
	workConn, err := ctl.connectServer()
	if err != nil {
		xl.Warn("无法和服务器节点建立链接: %v", err)
		return
	}

	m := &msg.NewWorkConn{
		RunID: ctl.runID,
	}
	if err = ctl.authSetter.SetNewWorkConn(m); err != nil {
		xl.Warn("验证新的工作链接时出错: %v", err)
		return
	}
	if err = msg.WriteMsg(workConn, m); err != nil {
		xl.Warn("无法将一个工作的链接写入至服务器: %v", err)
		workConn.Close()
		return
	}

	var startMsg msg.StartWorkConn
	if err = msg.ReadMsgInto(workConn, &startMsg); err != nil {
		xl.Trace("工作链接在服务器返回开始工作链接前关闭: %v", err)
		workConn.Close()
		return
	}
	if startMsg.Error != "" {
		xl.Error("开始工作线程通知匹配失败: %s", startMsg.Error)
		workConn.Close()
		return
	}

	// dispatch this work connection to related proxy
	ctl.pm.HandleWorkConn(startMsg.ProxyName, workConn, &startMsg)
}

func (ctl *Control) HandleNewProxyResp(inMsg *msg.NewProxyResp) {
	xl := ctl.xl
	// Server will return NewProxyResp message to each NewProxy message.
	// Start a new proxy handler if no error got
	err := ctl.pm.StartProxy(inMsg.ProxyName, inMsg.RemoteAddr, inMsg.Error)
	if err != nil {
		xl.Warn("[%s] 启动失败: %v", inMsg.ProxyName, err)
	} else {
		// http(s) 隧道规则
		if ctl.pxyCfgs[inMsg.ProxyName].GetBaseConfig().ProxyType == "https" || ctl.pxyCfgs[inMsg.ProxyName].GetBaseConfig().ProxyType == "http" {
			xl.Info("[%s] 映射启动成功, 感谢您使用LCF!", inMsg.ProxyName)
			xl.Info("您可以使用 [%s] 连接至您的隧道: %s", inMsg.RemoteAddr, strings.Split(inMsg.ProxyName, ".")[1])
		} else {
			xl.Info("[%s] 映射启动成功, 感谢您使用LCF!", inMsg.ProxyName)
			xl.Info("您可以使用 [%s] 连接至您的隧道: %s", ctl.clientCfg.ServerAddr+inMsg.RemoteAddr, strings.Split(inMsg.ProxyName, ".")[1])
		}
	}
}

func (ctl *Control) HandleNatHoleResp(inMsg *msg.NatHoleResp) {
	xl := ctl.xl

	// Dispatch the NatHoleResp message to the related proxy.
	ok := ctl.msgTransporter.DispatchWithType(inMsg, msg.TypeNameNatHoleResp, inMsg.TransactionID)
	if !ok {
		xl.Trace("dispatch NatHoleResp message to related proxy error")
	}
}

func (ctl *Control) Close() error {
	return ctl.GracefulClose(0)
}

func (ctl *Control) GracefulClose(d time.Duration) error {
	ctl.pm.Close()
	ctl.vm.Close()

	time.Sleep(d)

	ctl.conn.Close()
	ctl.cm.Close()
	return nil
}

// ClosedDoneCh returns a channel that will be closed after all resources are released
func (ctl *Control) ClosedDoneCh() <-chan struct{} {
	return ctl.closedDoneCh
}

// connectServer return a new connection to frps
func (ctl *Control) connectServer() (conn net.Conn, err error) {
	return ctl.cm.Connect()
}

// reader read all messages from frps and send to readCh
func (ctl *Control) reader() {
	xl := ctl.xl
	defer func() {
		if err := recover(); err != nil {
			xl.Error("panic 运行过程中出错: %v", err)
			xl.Error(string(debug.Stack()))
		}
	}()
	defer ctl.readerShutdown.Done()
	defer close(ctl.closedCh)

	encReader := crypto.NewReader(ctl.conn, []byte(ctl.clientCfg.Token))
	for {
		m, err := msg.ReadMsg(encReader)
		if err != nil {
			if err == io.EOF {
				xl.Debug("服务器返回空链接")
				return
			}
			xl.Warn("读取返回内容失败: %v", err)
			ctl.conn.Close()
			return
		}
		ctl.readCh <- m
	}
}

// writer writes messages got from sendCh to frps
func (ctl *Control) writer() {
	xl := ctl.xl
	defer ctl.writerShutdown.Done()
	encWriter, err := crypto.NewWriter(ctl.conn, []byte(ctl.clientCfg.Token))
	if err != nil {
		xl.Error("加密新的 Writer 失败: %v", err)
		ctl.conn.Close()
		return
	}
	for {
		m, ok := <-ctl.sendCh
		if !ok {
			xl.Info("正在和服务器断开链接...")
			return
		}

		if err := msg.WriteMsg(encWriter, m); err != nil {
			xl.Warn("将信息写入至工作链接时出错: %v", err)
			return
		}
	}
}

// msgHandler handles all channel events and performs corresponding operations.
func (ctl *Control) msgHandler() {
	xl := ctl.xl
	defer func() {
		if err := recover(); err != nil {
			xl.Error("panic 运行过程中出错: %v", err)
			xl.Error(string(debug.Stack()))
		}
	}()
	defer ctl.msgHandlerShutdown.Done()

	var hbSendCh <-chan time.Time
	// TODO(fatedier): disable heartbeat if TCPMux is enabled.
	// Just keep it here to keep compatible with old version frps.
	if ctl.clientCfg.HeartbeatInterval > 0 {
		hbSend := time.NewTicker(time.Duration(ctl.clientCfg.HeartbeatInterval) * time.Second)
		defer hbSend.Stop()
		hbSendCh = hbSend.C
	}

	var hbCheckCh <-chan time.Time
	// Check heartbeat timeout only if TCPMux is not enabled and users don't disable heartbeat feature.
	if ctl.clientCfg.HeartbeatInterval > 0 && ctl.clientCfg.HeartbeatTimeout > 0 && !ctl.clientCfg.TCPMux {
		hbCheck := time.NewTicker(time.Second)
		defer hbCheck.Stop()
		hbCheckCh = hbCheck.C
	}

	ctl.lastPong = time.Now()
	for {
		select {
		case <-hbSendCh:
			// send heartbeat to server
			xl.Debug("向服务器发送心跳包")
			pingMsg := &msg.Ping{}
			if err := ctl.authSetter.SetPing(pingMsg); err != nil {
				xl.Warn("error during ping authentication: %v", err)
				return
			}
			ctl.sendCh <- pingMsg
		case <-hbCheckCh:
			if time.Since(ctl.lastPong) > time.Duration(ctl.clientCfg.HeartbeatTimeout)*time.Second {
				xl.Warn("heartbeat timeout")
				// let reader() stop
				ctl.conn.Close()
				return
			}
		case rawMsg, ok := <-ctl.readCh:
			if !ok {
				return
			}

			switch m := rawMsg.(type) {
			case *msg.ReqWorkConn:
				go ctl.HandleReqWorkConn(m)
			case *msg.NewProxyResp:
				ctl.HandleNewProxyResp(m)
			case *msg.NatHoleResp:
				ctl.HandleNatHoleResp(m)
			case *msg.CloseClient:
				ctl.HandleCloseClient(m)
			case *msg.Pong:
				if m.Error != "" {
					xl.Error("Pong contains error: %s", m.Error)
					ctl.conn.Close()
					return
				}
				ctl.lastPong = time.Now()
				xl.Debug("receive heartbeat from server")
			}
		}
	}
}

func (ctl *Control) HandleCloseClient(ClientInfo *msg.CloseClient) {
	xl := ctl.xl
	if ClientInfo.Token != ctl.clientCfg.User {
		xl.Warn("客户端 User 与传递的 User 不符")
		return
	}
	xl.Info("您的客户端已被服务端强制下线")
	os.Exit(0)
}

// If controler is notified by closedCh, reader and writer and handler will exit
func (ctl *Control) worker() {
	go ctl.msgHandler()
	go ctl.reader()
	go ctl.writer()

	<-ctl.closedCh
	// close related channels and wait until other goroutines done
	close(ctl.readCh)
	ctl.readerShutdown.WaitDone()
	ctl.msgHandlerShutdown.WaitDone()

	close(ctl.sendCh)
	ctl.writerShutdown.WaitDone()

	ctl.pm.Close()
	ctl.vm.Close()

	close(ctl.closedDoneCh)
	ctl.cm.Close()
}

func (ctl *Control) ReloadConf(pxyCfgs map[string]config.ProxyConf, visitorCfgs map[string]config.VisitorConf) error {
	ctl.vm.Reload(visitorCfgs)
	ctl.pm.Reload(pxyCfgs)
	return nil
}
