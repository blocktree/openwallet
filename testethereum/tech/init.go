package tech

import "github.com/blocktree/openwallet/manager"

var (
	tm      = openw.NewWalletManager(NewEthTestConfig())
	testApp = "openw"
)

func init() {
	//	tm.Init()
}
