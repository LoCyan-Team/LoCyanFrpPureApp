# LoCyanFrpPureApp

基于原版 Frp 的修改版本，原仓库：  
<https://github.com/fatedier/frp>

## 修改内容

### 服务端

- 从远程 API 获取限速信息
- 从远程 API 校验访问密钥
- 从远程 API 校验隧道权限
- 添加了 Run ID 数据上报
- 添加节点 `node_id` 配置项
- 移除了 `api_baseurl` 配置项

### 客户端

- 添加了简易启动命令
