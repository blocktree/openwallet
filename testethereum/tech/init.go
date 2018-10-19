package tech

import "github.com/blocktree/OpenWallet/manager"

var (
	tm      = manager.NewWalletManager(NewEthTestConfig())
	testApp = "openw"
)

func init() {
	//	tm.Init()
}
