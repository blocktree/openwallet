package tech

import (
	"os"

	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/assets/stc2345"
	"github.com/blocktree/OpenWallet/log"
)

func Getetc2345WalletManager() (*stc2345.WalletManager, error) {
	manager, ok := assets.Managers["stc2345"]
	if !ok {
		log.Error("cannot find the stc2345 manager.")
		os.Exit(-1)
	}
	return manager.(*stc2345.WalletManager), nil
}

func TestCreateWallet2345() {
	manager, _ := Getetc2345WalletManager()
	err := manager.CreateWalletFlow()
	if err != nil {
		log.Error("create wallet failed, err=", err)
	}
}

func TestBatchCreateAddr2345() {
	manager, _ := Getetc2345WalletManager()
	err := manager.CreateAddressFlow()
	if err != nil {
		log.Error("create wallet failed, err=", err)
	}
}
