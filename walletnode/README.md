## Walletnode
--------------

Openwallet 基础架构中，关于 Wallet Fullnode 管理相关的接口。含两个部分：
  - `wmd node XXX -s Symbol` 操作
	- Golang API for walletnode managment(Like: start/stop/restart/copy)

### 其中 Golang 接口调用示例一：启动，重启，关闭
```
	import "github.com/blocktree/OpenWallet/walletnode"

	symbol := "bopo" // 币种，同 assets.Symbol
	wn := walletnode.WalletnodeManager{}

	// 关闭钱包节点， return error
	if err := wn.StopNodeFlow(symbol); err != nil {
		log.Println(err)
	}

	// 开启钱包节点， return error
	if err := wn.StartNodeFlow(symbol); err != nil {
		log.Println(err)
	}

	// 重启钱包节点， return error
	if err := wn.RestartNodeFlow(symbol); err != nil {
		log.Println(err)
	}
```
--------------
### Golang 接口调用示例二：备份/恢复(only files)
```
import "github.com/blocktree/OpenWallet/walletnode"

symbol := "bopo"
wn := walletnode.WalletnodeManager{}
```

### wmd node 操作相关

如果使用 `docker+自制镜像` 作为钱包节点（无论docker是在本地还是远程），都需要先执行 `wmd node create -s Symbol`， 否则跳过。过程：

（[自制镜像](http://192.168.5.138:3000/WalletTeam/WalletImageRegistry)）

	先执行  wmd node create -s Symbol
	后执行  wmd wallet create -s Symbol
	最后：任何 wmd 命令都可用

wmd node ceate -s Symbol 操作的结果，以 conf/SYMBOL.ini 文件内容的形式输出，这里无专门的 API。

任何对 walletnode 的数据需求，都可以通过此 ini 文件获得（比如 rpcUser/rpcPassword/rpcURL/httpURL 等）



Done!

============================================ 以下内容适合深度了解

## 本接口功能定位


正常情形下，OpenWallet 创建一个币的全节点，含：
  1. 创建节点  wmd node create -s BCH
  2. 创建钱包  wmd wallet create -s BCH
  3. ... 其他业务操作（transfer, backup, restore...)

本接口完成`第一步`节点管理相关操作，有：
  1. 创建 wmd node create ...
  2. 启动 wmd node start ...
  3. 停止 wmd node stop ...
  4. 重启 wmd node restart ...
  5. 运行状态查看 wmd node status ...

其中创建节点的具体内容，涉：
  1. 创建 conf/SYMBOL.ini 文件 （注意：与 wmd wallet config 协同）
  2. 写入 Walletnode 相关参数 (Walletnode 服务器 类型/地址/文件描述符/前缀...)

## 本接口的作用

由于 BTC 等钱包中，在恢复钱包，导入私钥等操作中，需要重启节点，所以定义了如下接口来实现节点的操作。

## 应用中的问题

有两个地方的代码需要升级：
  1. 之前采用指定启动命令方式来重启钱包的，请升级为如下接口来实现
  2. 之前采用本地目录copy方式备份和恢复的，请升级为如下复制接口来实现

其他问题随时 @luo

## 其他技术细节

  1. 为了兼容开发，测试，生产环境中 walletnode 不同的部署方式，通过 .ini 文件中 `WalletnodeServerType=service/localdocker/remotedocker` 三个参数来指定全节点是 直接安装在裸机或本地PC/安装在本地Docker/安装在远程服务器的Docker。指定后，接口中将自动处理后续问题（连接，pull镜像，创建，备份，恢复等）。


  2. 同样，备份也有上述需求（本地的 copy，或 远程的网络传输），通过 WalletnodeManager 这个 Interface 实现几个方法，解决：
	- 自动选择是从本地还是Docker中备份/恢复文件，避免开发时采用本地 cp，而生产中需要 docker cp（经过 docker 处理备份）的冲突

