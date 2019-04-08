#OpenWallet

## Build development environment

The requirements to build OpenWallet are:

- Golang version 1.10 or later
- govendor (a third party package management tool)
- xgo (Go CGO cross compiler)
- Properly configured Go language environment
- Golang supported operating system

## 依赖库管理工具govendor

### 安装govendor

```shell

go get -u -v github.com/kardianos/govendor

```

### 使用govendor

```shell

#进入到项目目录
$ cd $GOPATH/src/github.com/blocktree/OpenWallet

#初始化vendor目录
$ govendor init

#查看vendor目录
[root@CC54425A openwallet]# ls
commands  main.go  vendor

#将GOPATH中本工程使用到的依赖包自动移动到vendor目录中
#说明：如果本地GOPATH没有依赖包，先go get相应的依赖包
$ govendor add +external
或使用缩写： govendor add +e

#Go 1.6以上版本默认开启 GO15VENDOREXPERIMENT 环境变量，可忽略该步骤。
#通过设置环境变量 GO15VENDOREXPERIMENT=1 使用vendor文件夹构建文件。
#可以选择 export GO15VENDOREXPERIMENT=1 或 GO15VENDOREXPERIMENT=1 go build 执行编译
$ export GO15VENDOREXPERIMENT=1

# 如果$GOPATH下已更新本地库，可执行命令以下命令，同步更新vendor包下的库
# 例如本地的$GOPATH/github.com/blocktree/下的组织项目更新后，可执行下面命令同步更新vendor
$ govendor update +v

```

## 源码编译跨平台工具

### 安装xgo（支持跨平台编译C代码）

[官方github](https://github.com/karalabe/xgo)

xgo的使用依赖docker。并且把要跨平台编译的项目文件加入到File sharing。

```shell

$ go get github.com/karalabe/xgo
...
$ xgo -h
...

```

## wmd--多币种钱包维护工具(Deprecated, please use go-openw-cli)

### 特点

wmd为了实现对多币种的钱包操作，规范了以下接口：

- 初始币种配置流程。
- 创建币种钱包流程。
- 批量币种钱包地址流程。
- 备份钱包流程。
- 启动定时器汇总钱包流程。

### 编译wmd工具

```shell

# 进入目录
$ $GOPATH/src/github.com/blocktree/OpenWallet/cmd/wmd

# 全部平台版本编译
$ xgo .

# 或自编译某个系统的版本
$ xgo --targets=linux/amd64 .

```

### wmd工具使用

wmd是一款多币种钱包维护工具。你只需要在服务器安装某币种的官方全节点钱包，并且wmd已经支持的币种。
你就可以使用wmd的规范的命令完成钱包维护工作。

#### 节点相关

```shell

# 自动安装[symbol]的官方节点，到[dir]目录。
$ ./wmd node install -s [symbol] -p [dir]

# 执行来自配置文件启动[symbol]节点的命令
$ ./wmd node start -s [symbol]

# 执行来自配置文件关闭[symbol]节点的命令
$ ./wmd node stop -s [symbol]

# 执行来自配置文件关闭[symbol]节点和启动[symbol]节点的命令
$ ./wmd node restart -s [symbol]

# 查看与[symbol]节点相关的信息
$ ./wmd node info -s [symbol]

# 查看./conf/[symbol].ini文件中与节点相关的配置信息
$ ./wmd node config -s [symbol]

# 执行重新初始化节点配置
$ ./wmd node config -s [symbol] -i

```

#### 钱包相关

```shell

# 创建钱包，成功后，文件保存在./data/[symbol]/key/
$ ./wmd wallet new -s [symbol]

# 备份钱包私钥和账户相关文件，文件保存在./data/[symbol]/key/backup/
$ ./wmd wallet backup -s [symbol]

# 执行恢复钱包，提供钱包的备份文件
$ ./wmd wallet restore -s [symbol]

# 执行批量创建地址命令，文件保存在./conf/[symbol]/address/
$ ./wmd wallet batchaddr -s [symbol]

# 启动批量汇总监听器
$ ./wmd wallet startsum -s [symbol]

# 查询钱包列表
$ ./wmd wallet list -s [symbol]

# 发起转行交易
$ ./wmd wallet transfer -s [symbol]

# 查看./conf/[symbol].ini文件中与钱包相关的配置信息
$ ./wmd wallet config -s [symbol]

# 执行重新初始化钱包配置
$ ./wmd wallet config -s [symbol] -i

```


