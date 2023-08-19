// Copyright 2018 fatedier, fatedier@gmail.com
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

package sub

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/sourcegraph/conc"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/api"
	"github.com/fatedier/frp/pkg/auth"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/version"
)

const (
	CfgFileTypeIni = iota
	CfgFileTypeCmd
)

var (
	cfgFile     string
	cfgDir      string
	cfgProxyIDs []string
	cfgToken    string
	showVersion bool

	serverAddr      string
	user            string
	protocol        string
	token           string
	logLevel        string
	logFile         string
	logMaxDays      int
	disableLogColor bool
	dnsServer       string

	proxyName          string
	localIP            string
	localPort          int
	remotePort         int
	useEncryption      bool
	useCompression     bool
	bandwidthLimit     string
	bandwidthLimitMode string
	customDomains      string
	subDomain          string
	httpUser           string
	httpPwd            string
	locations          string
	hostHeaderRewrite  string
	role               string
	sk                 string
	multiplexer        string
	serverName         string
	bindAddr           string
	bindPort           int

	tlsEnable     bool
	tlsServerName string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./frpc.ini", "config file of frpc")
	rootCmd.PersistentFlags().StringVarP(&cfgDir, "config_dir", "", "", "config directory, run one frpc service for each file in config directory")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of frpc")
	rootCmd.PersistentFlags().StringVarP(&cfgToken, "token", "u", "", "The Token of LoCyanFrp")
	rootCmd.PersistentFlags().StringSliceVarP(&cfgProxyIDs, "id", "p", nil, "The ProxyID of LoCyanFrp")
}

func RegisterCommonFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&serverAddr, "server_addr", "s", "127.0.0.1:7000", "frp server's address")
	cmd.PersistentFlags().StringVarP(&user, "user", "u", "", "user")
	cmd.PersistentFlags().StringVarP(&protocol, "protocol", "p", "tcp", "tcp, kcp, quic, websocket, wss")
	cmd.PersistentFlags().StringVarP(&token, "token", "t", "", "auth token")
	cmd.PersistentFlags().StringVarP(&logLevel, "log_level", "", "info", "log level")
	cmd.PersistentFlags().StringVarP(&logFile, "log_file", "", "console", "console or file path")
	cmd.PersistentFlags().IntVarP(&logMaxDays, "log_max_days", "", 3, "log file reversed days")
	cmd.PersistentFlags().BoolVarP(&disableLogColor, "disable_log_color", "", false, "disable log color in console")
	cmd.PersistentFlags().BoolVarP(&tlsEnable, "tls_enable", "", true, "enable frpc tls")
	cmd.PersistentFlags().StringVarP(&tlsServerName, "tls_server_name", "", "", "specify the custom server name of tls certificate")
	cmd.PersistentFlags().StringVarP(&dnsServer, "dns_server", "", "", "specify dns server instead of using system default one")
}

var rootCmd = &cobra.Command{
	Use:   "frpc",
	Short: "Edited from fatedier/frp, Powered by LoCyanTeam",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}

		log.Info("欢迎使用LoCyanFrp映射客户端! v0.51.3 #20230710001")

		// If cfgDir is not empty, run multiple frpc service for each config file in cfgDir.
		// Note that it's only designed for testing. It's not guaranteed to be stable.
		if cfgDir != "" {
			_ = runMultipleClients(cfgDir)
			return nil
		}

		log.Info("To Get Config File from LoCyanFrp API...")
		s, err := api.NewService("https://www.locyanfrp.cn/api/")

		if err != nil {
			log.Warn("Initialize API Service Failed, err: %s", err)
		}

		if cfgToken != "" && len(cfgProxyIDs) > 0 {
			if len(cfgProxyIDs) > 1 {
				var wg conc.WaitGroup

				// 每一个都是新的协程
				for _, proxyID := range cfgProxyIDs {
					// 将循环变量赋值给局部变量
					proxyID := proxyID

					wg.Go(func() {
						configPath := filepath.Join("ini", fmt.Sprintf("%s.ini", proxyID))

						configContent, err := s.EZStartGetCfg(cfgToken, proxyID)
						if err != nil {
							// 无法获取配置文件，直接关闭软件，防止启动上一个配置文件导致二次报错
							log.Error("Get config file failed, please check your arguments: %s", err)

							os.Exit(1)
						}

						log.Debug("Loading proxy config file: %s", configPath)

						configFile, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
						if err != nil {
							log.Error("Open config file failed: %s", err)

							os.Exit(1)
						}

						defer configFile.Close()

						_, err = configFile.WriteString(configContent) // 直接写入字符串数据
						// 写入文件是否成功检测
						if err != nil {
							log.Error("Write config failed: %s", err)

							os.Exit(1)
						}

						// 内容写入后直接启动
						if err := runClient(configPath); err != nil {
							log.Error("Run client failed: %s", err)

							os.Exit(1)
						}
					})
				}

				wg.Wait()

				return nil
			}

			// 没有多开现象
			cfg, err := s.EZStartGetCfg(cfgToken, cfgProxyIDs[0])
			if err != nil {
				log.Warn("Get config file failed, please check your arguments: %s", err)
				// 无法获取配置文件，直接关闭软件，防止启动上一个配置文件导致二次报错
				os.Exit(1)
			}

			// 删除原先文件，防止窜行
			// OpenFlie 函数用 os.O_RDWR|os.O_TRUNC|os.O_CREATE 可以直接覆盖原文本
			// os.RemoveAll("./frpc.ini")

			// 有则打开，无则新建
			file, err := os.OpenFile("./frpc.ini", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			if err != nil {
				log.Warn("Open frpc.ini failed: %s", err)

				os.Exit(1)
			}

			defer file.Close()

			if _, err := file.WriteString(cfg); err != nil {
				log.Error("Cant write content to config file: %s", err)

				os.Exit(1)
			}

			// 内容写入后直接启动
			if err := runClient(cfgFile); err != nil {
				log.Error("Run client failed: %s", err)

				os.Exit(1)
			}

			return nil
		}

		// Do not show command usage here.
		if err := runClient(cfgFile); err != nil {
			os.Exit(1)
		}

		return nil
	},
}

func runMultipleClients(cfgDir string) error {
	var wg conc.WaitGroup
	defer wg.Wait()
	err := filepath.WalkDir(cfgDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		time.Sleep(time.Millisecond)
		wg.Go(func() {
			err := runClient(path)
			if err != nil {
				fmt.Printf("frpc service error for config file [%s]\n", path)
			}
		})
		return nil
	})
	return err
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func handleTermSignal(svr *client.Service) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	svr.GracefulClose(500 * time.Millisecond)
}

func parseClientCommonCfgFromCmd() (cfg config.ClientCommonConf, err error) {
	cfg = config.GetDefaultClientConf()

	ipStr, portStr, err := net.SplitHostPort(serverAddr)
	if err != nil {
		err = fmt.Errorf("invalid server_addr: %v", err)
		return
	}

	cfg.ServerAddr = ipStr
	cfg.ServerPort, err = strconv.Atoi(portStr)
	if err != nil {
		err = fmt.Errorf("invalid server_addr: %v", err)
		return
	}

	cfg.User = user
	cfg.Protocol = protocol
	cfg.LogLevel = logLevel
	cfg.LogFile = logFile
	cfg.LogMaxDays = int64(logMaxDays)
	cfg.DisableLogColor = disableLogColor
	cfg.DNSServer = dnsServer

	// Only token authentication is supported in cmd mode
	cfg.ClientConfig = auth.GetDefaultClientConf()
	cfg.Token = token
	cfg.TLSEnable = tlsEnable
	cfg.TLSServerName = tlsServerName

	cfg.Complete()
	if err = cfg.Validate(); err != nil {
		err = fmt.Errorf("parse config error: %v", err)
		return
	}
	return
}

func runClient(cfgFilePath string) error {
	cfg, pxyCfgs, visitorCfgs, err := config.ParseClientConfig(cfgFilePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return startService(cfg, pxyCfgs, visitorCfgs, cfgFilePath)
}

func startService(
	cfg config.ClientCommonConf,
	pxyCfgs map[string]config.ProxyConf,
	visitorCfgs map[string]config.VisitorConf,
	cfgFile string,
) (err error) {
	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel,
		cfg.LogMaxDays, cfg.DisableLogColor)

	if cfgFile != "" {
		log.Info("start frpc service for config file [%s]", cfgFile)
		defer log.Info("frpc service for config file [%s] stopped", cfgFile)
	}
	svr, errRet := client.NewService(cfg, pxyCfgs, visitorCfgs, cfgFile)
	if errRet != nil {
		err = errRet
		return
	}

	shouldGracefulClose := cfg.Protocol == "kcp" || cfg.Protocol == "quic"
	// Capture the exit signal if we use kcp or quic.
	if shouldGracefulClose {
		go handleTermSignal(svr)
	}

	_ = svr.Run(context.Background())
	return
}
