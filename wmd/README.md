# wmd

## Build development environment

The requirements to build openwallet are:

- Golang version 1.12 or later
- xgo (Go CGO cross compiler)
- Properly configured Go language environment
- Golang supported operating system

## 源码编译跨平台工具

### 安装xgo（支持跨平台编译C代码）

[官方github（目前还不支持go module）](https://github.com/karalabe/xgo)
[支持go module的xgo fork](https://github.com/gythialy/xgo)

xgo的使用依赖docker。并且把要跨平台编译的项目文件加入到File sharing。

```shell

# 官方的目前还不支持go module编译，所以我们用了别人改造后能够给支持的fork版
$ go get -u github.com/gythialy/xgo
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

# 本地系统编译
$ make clean build

# 跨平台编译wmd，更多可选平台可修改Makefile的$TARGETS变量
$ make clean wmd

```

### wmd工具使用

wmd是一款多币种钱包维护工具。你只需要在服务器安装某币种的官方全节点钱包，并且wmd已经支持的币种。
你就可以使用wmd的规范的命令完成钱包维护工作。

#### 钱包相关

```shell

# 查看./conf/[symbol].ini文件中与钱包相关的配置信息
$ ./wmd wallet config -s [symbol]

# 执行重新初始化钱包配置
$ ./wmd wallet config -s [symbol] -i

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

```