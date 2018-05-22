#OpenWallet

## 依赖库管理工具govendor

### 安装govendor

```shell

go get -u -v github.com/kardianos/govendor

```

### 使用govendor

```shell

#进入到项目目录
$ cd $GOPATH/OpenWallet

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

```

## 源码编译跨平台工具

### 安装gox（无法跨平台编译C代码，弃用）

```shell

$ go get github.com/mitchellh/gox
...
$ gox -h
...

```

### 安装xgo（支持跨平台编译C代码）

[官方github](https://github.com/karalabe/xgo)

xgo的使用依赖docker。并且把要跨平台编译的项目文件加入到File sharing。

```shell

$ go get github.com/karalabe/xgo
...
$ xgo -h
...

```

## wmd--多币种钱包维护工具

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
$ $GOPATH/OpenWallet/cmd/wmd

# 全部平台版本编译
$ xgo .

# 或自编译某个系统的版本
$ xgo --targets=linux/amd64 .

```

### wmd工具使用

wmd是一款多币种钱包维护工具。你只需要在服务器安装某币种的官方全节点钱包，并且wmd已经支持的币种。
你就可以使用wmd的规范的命令完成钱包维护工作。

```shell

# 上传wmd文件到你的钱包服务器

# 命令行中： -s <symbol> 是针对某个币

#执行初始化配置文件，文件保存在./conf/<symbol>.json
$ ./wmd config init -s <symbol>

#执行查看钱包管理工具的配置文件
$ ./wmd config see -s <symbol>

#创建钱包，成功后，文件保存在./data/<symbol>/key/
$ ./wmd wallet new -s <symbol>

#备份钱包私钥和账户相关文件，文件保存在./data/<symbol>/key/backup/
$ ./wmd wallet backup -s <symbol>

#执行批量创建地址命令，文件保存在./conf/<symbol>/address/
$ ./wmd wallet batchaddr -s <symbol>

#启动批量汇总监听器
$ ./wmd wallet startsum -s <symbol>

```