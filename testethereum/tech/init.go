package tech

import "github.com/blocktree/OpenWallet/manager"

var (
	tm      = openw.NewWalletManager(NewEthTestConfig())
	testApp = "openw"
)

func init() {
	//	tm.Init()
}
