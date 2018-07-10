package main

import (
	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/openwallet"
)

func main() {

	//添加eth进入资产路由中
	eth := &assets.EthereumAssets{}
	openwallet.Router(eth.Name(), eth)

	//运行应用
	openwallet.Run()

}
