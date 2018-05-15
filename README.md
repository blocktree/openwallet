#OpenWallet

## 依赖库管理工具govendor

### 安装govendor

```shell

go get -u -v github.com/kardianos/govendor

```

### 使用govendor

```shell

#进入到项目目录
cd $GOPATH/openwallet

#初始化vendor目录
govendor init

#查看vendor目录
[root@CC54425A openwallet]# ls
commands  main.go  vendor

#将GOPATH中本工程使用到的依赖包自动移动到vendor目录中
#说明：如果本地GOPATH没有依赖包，先go get相应的依赖包
govendor add +external
或使用缩写： govendor add +e

#Go 1.6以上版本默认开启 GO15VENDOREXPERIMENT 环境变量，可忽略该步骤。
#通过设置环境变量 GO15VENDOREXPERIMENT=1 使用vendor文件夹构建文件。
#可以选择 export GO15VENDOREXPERIMENT=1 或 GO15VENDOREXPERIMENT=1 go build 执行编译
export GO15VENDOREXPERIMENT=1

```