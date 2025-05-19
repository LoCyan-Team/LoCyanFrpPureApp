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
	"github.com/fatedier/frp/pkg/api"
	"github.com/fatedier/frp/pkg/api/client/tunnel"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/config"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/featuregate"
	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/version"
)

var (
	cfgFile          string
	cfgDir           string
	showVersion      bool
	strictConfigMode bool
	//quickStart       string
	lcfFrpToken  string
	lcfTunnelIds []int64
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./config.json", "指定 Frp 客户端配置文件")
	rootCmd.PersistentFlags().StringVarP(&cfgDir, "config_dir", "", "", "指定配置文件夹，一个文件将运行一个 Frp 客户端服务")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Frp 客户端版本")
	rootCmd.PersistentFlags().BoolVarP(&strictConfigMode, "strict_config", "", true, "严格配置解析模式，未知配置将产生错误")
	//rootCmd.PersistentFlags().StringVarP(&quickStart, "start", "s", "", "LoCyanFrp 快速启动隧道")
	rootCmd.PersistentFlags().StringVarP(&lcfFrpToken, "token", "u", "", "LoCyanFrp 用户访问令牌")
	rootCmd.PersistentFlags().Int64SliceVarP(&lcfTunnelIds, "id", "p", []int64{}, "LoCyanFrp 隧道 ID 列表")
}

var rootCmd = &cobra.Command{
	Use:   "frpc",
	Short: "frpc 是 Frp 的客户端，当前程序由乐青映射定制 (https://github.com/LoCyan-Team/LoCyanFrpPureApp)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}
		log.Infof("欢迎使用 LoCyanFrp 客户端")

		if lcfFrpToken != "" && len(lcfTunnelIds) > 0 {
			err := quickStartClient(lcfFrpToken, lcfTunnelIds)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return nil
		}

		// If cfgDir is not empty, run multiple frpc service for each config file in cfgDir.
		// Note that it's only designed for testing. It's not guaranteed to be stable.
		if cfgDir != "" {
			err := runMultipleClients(cfgDir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return nil
		}

		// Do not show command usage here.
		err := runClient(cfgFile)
		if err != nil {
			log.Errorf("启动配置 [%s] 出错: %v", cfgFile, err)
			os.Exit(1)
		}
		return nil
	},
}

func runMultipleClients(cfgDir string) error {
	var wg sync.WaitGroup
	err := filepath.WalkDir(cfgDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		wg.Add(1)
		time.Sleep(time.Millisecond)
		go func() {
			defer wg.Done()
			err := runClient(path)
			if err != nil {
				log.Warnf("Frp 客户端配置 [%s] 启动错误：%s", path, err)
			}
		}()
		return nil
	})
	wg.Wait()
	return err
}

func Execute() {
	rootCmd.SetGlobalNormalizationFunc(config.WordSepNormalizeFunc)
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

func runClient(cfgFilePath string) error {
	cfg, proxyCfgs, visitorCfgs, isLegacyFormat, err := config.LoadClientConfig(cfgFilePath, strictConfigMode)
	if err != nil {
		return err
	}
	if isLegacyFormat {
		fmt.Printf("警告: INI 格式已弃用，将于未来版本的 Frp 移除，" +
			"请使用 JSON/TOML/YAML 格式配置代替！\n")
	}

	if len(cfg.FeatureGates) > 0 {
		if err := featuregate.SetFromMap(cfg.FeatureGates); err != nil {
			return err
		}
	}

	warning, err := validation.ValidateAllClientConfig(cfg, proxyCfgs, visitorCfgs)
	if warning != nil {
		fmt.Printf("警告: %v\n", warning)
	}
	if err != nil {
		return err
	}
	return startService(cfg, proxyCfgs, visitorCfgs, cfgFilePath)
}

func startService(
	cfg *v1.ClientCommonConfig,
	proxyCfgs []v1.ProxyConfigurer,
	visitorCfgs []v1.VisitorConfigurer,
	cfgFile string,
) error {
	log.InitLogger(cfg.Log.To, cfg.Log.Level, int(cfg.Log.MaxDays), cfg.Log.DisablePrintColor)

	if cfgFile != "" {
		log.Infof("为配置 [%s] 启动服务", cfgFile)
		defer log.Infof("配置 [%s] 服务已停止", cfgFile)
	}
	svr, err := client.NewService(client.ServiceOptions{
		Common:         cfg,
		ProxyCfgs:      proxyCfgs,
		VisitorCfgs:    visitorCfgs,
		ConfigFilePath: cfgFile,
	})
	if err != nil {
		return err
	}

	shouldGracefulClose := cfg.Transport.Protocol == "kcp" || cfg.Transport.Protocol == "quic"
	// Capture the exit signal if we use kcp or quic.
	if shouldGracefulClose {
		go handleTermSignal(svr)
	}
	return svr.Run(context.Background())
}

// quickStartClient 一键启动
func quickStartClient(frpToken string, tunnelIds []int64) error {
	log.Infof("正在从 LoCyanFrp API 获取配置文件...")
	as, err := api.NewApiService()
	if err != nil {
		log.Warnf("初始化 API 服务失败")
		return err
	}
	var wg sync.WaitGroup

	// 新建配置文件文件夹
	cacheDir := ".lcf-cache"

	_, err = os.Stat(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(cacheDir, os.ModePerm)
			if err != nil {
				log.Errorf("创建缓存文件夹出错")
				return err
			}
		} else {
			log.Errorf("创建缓存文件夹出错")
			return err
		}
	}

	// 每一个都是新的协程
	for _, tunnelId := range tunnelIds {
		// 将循环变量赋值给局部变量
		currentTunnelId := tunnelId

		configPath := filepath.Join(cacheDir, fmt.Sprintf("%s.json", strconv.FormatInt(currentTunnelId, 10)))

		apiGetConfig, err := as.Client.Tunnel.GetConfig(tunnel.GetConfigParams{
			FrpToken: frpToken,
			TunnelId: currentTunnelId,
		})
		if err != nil {
			// 无法获取配置文件，直接关闭软件，防止启动上一个配置文件导致二次报错
			log.Errorf("获取隧道 [%d] 配置文件失败", tunnelId)
			return err
		}
		if apiGetConfig.Status != 200 {
			log.Errorf("获取隧道 [%d] 配置文件失败，API 返回消息: %s", tunnelId, apiGetConfig.Message)
			return nil
		}

		wg.Add(1)
		time.Sleep(time.Millisecond)
		go func(tunnelId int64, cfgPath string, jsonCfg string) { // 传递必要参数
			defer wg.Done()
			// 内部处理文件创建、写入和关闭
			if err := func() error {
				configFile, err := os.OpenFile(cfgPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
				if err != nil {
					log.Errorf("Frp 客户端隧道 [%d] 打开配置文件出错: %v", tunnelId, err)
					return err
				}
				defer configFile.Close()

				_, err = configFile.WriteString(jsonCfg)
				if err != nil {
					log.Errorf("Frp 客户端隧道 [%d] 写入配置文件出错: %v", tunnelId, err)
					return err
				}
				return nil
			}(); err != nil {
				return // 如果文件操作失败，直接退出
			}

			err := runClient(cfgPath)
			if err != nil {
				log.Errorf("Frp 客户端隧道 [%d] 启动出错: %v", tunnelId, err)
			}
		}(currentTunnelId, configPath, apiGetConfig.Data.Config) // 传递参数
	}

	wg.Wait()
	return nil
}
