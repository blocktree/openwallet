## Walletnode
--------------

Openwallet 基础架构中，生产环境使用 Docker 作为全节点钱包。这将使得重启，备份等操作，都需要实现 “远程”，也就是跨 Docker 容器的调用。

而，大多数开发环境，全节点都是直接安装在本地，这显然和生产环境不一样，那么在同样的代码中，如何使得开发和生产环境适配，是个“问题”。

本节实现基于 Docker 的 Fullnode 管理接口，应用于全节点的：
- start/stop/restart，重启节点(如导入私钥等)
- copy/back，远程备份和恢复(如：生产环境中使用docker，而不是本地安装来部署全节点，这时候显然需要跨容器远程复制，而不是本地 cp，因为数据在容器)全节点钱包数据（如：wallet.dat等文件）
- `兼容本地操作`的功能，已开发，待讨论决定

### 其中 Golang 接口调用示例一：启动，重启，关闭

```
	import "github.com/blocktree/OpenWallet/walletnode"

	symbol := "bopo" // 币种，同 assets.Symbo
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

// 备份
src := "/data/wallet.dat"  // 备份来源，全节点中的文件 (如： src = MainDataPath + '/' + filename)
dst := "/tmp/2018...../wallet.dat" // 备份目标，自设
if err := wn.CopyFromContainer(symbol, src, dst); err != nil {
	return err
}

// 恢复
src := "/tmp/2018....../wallet.dat"  // 恢复来源，用户提供
dst := "/data/wallet.dat" // 恢复目标的文件名 (如：dst = MainDataPath + '/')
if err := wn.CopyToContainer(symbol, src, dst); err != nil {
	return err
}

```

### 使用 wmd node 创建全节点

如果使用 `docker+自制镜像` 作为钱包节点（无论docker是在本地还是远程），都需要先执行 `wmd node create -s Symbol`， 否则跳过。过程：

（[自制镜像](http://192.168.5.138:3000/WalletTeam/WalletImageRegistry)）

	先执行  wmd node create -s Symbol
	后执行  wmd wallet create -s Symbol
	最后：任何 wmd 命令都可用

wmd node ceate -s Symbol 操作的结果，以 conf/SYMBOL.ini 文件内容的形式输出，这里无专门的 API。

任何对 walletnode 的数据需求，都可以通过此 ini 文件获得（比如 rpcUser/rpcPassword/rpcURL/httpURL 等）

```
simonluo@MBP15L:mainnet/$ wmd node create -s qtum
2018/09/10 20:18:50 [N] Wallet Manager Load Successfully.
Config file <QTUM.ini> existed!
Init new QTUM wallet fullnode in '/Users/simonluo/.wmd/mainnet/'(
  yes:   to create config file and docker,
  no:    just to create docker,
[yes]: yes																								// 已存在 ini 文件，选择是否重写，否则创建一个新的
Within testnet('testnet','main')[testnet]: main           // 主网/测试链
Where to run Walletnode: local/docker [docker]: docker    // 使用容器/本地安装的方式部署全节点
Docker master server addr [192.168.2.194]: 192.168.2.194  // 如果选择 docker，需提供 master 的 IP 地址
Docker master server port [2375]: 2375                    // Docker 服务端口
Start to create/update config file...
         create success!
         update success!
QTUM walletnode exist: running
simonluo@MBP15L:mainnet/$
```

如果选择本地安装的全节点（一般用来自己测试或开发）
```
Within testnet('testnet','main')[testnet]:
Where to run Walletnode: local/docker [docker]: local
Start walletnode command: /usr/local/bin/bitcoin-cli XXXX       // 输入启动命令
Stop walletnode command: /usr/local/bin/bitcoin-cli XXXX stop    // 输入关闭命令
Start to create/update config file...
         create success!
         update success!
```

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

  1. 为了兼容开发，测试，生产环境中 walletnode 不同的部署方式，通过 .ini 文件中 `ServerType=service/localdocker/remotedocker` 三个参数来指定全节点是 直接安装在裸机或本地PC/安装在本地Docker/安装在远程服务器的Docker。指定后，接口中将自动处理后续问题（连接，pull镜像，创建，备份，恢复等）


  2. 同样，备份也有上述需求（本地的 copy，或 远程的网络传输），通过 WalletnodeManager 这个 Interface 实现几个方法，解决：
	 - 自动选择是从本地还是Docker中备份/恢复文件，避免开发时采用本地 cp，而生产中需要 docker cp（经过 docker 处理备份）的冲突

