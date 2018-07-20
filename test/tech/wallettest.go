package tech

import (
	"fmt"
	"math/big"

	"github.com/blocktree/OpenWallet/assets/ethereum"
)

func TestNewWallet(aliaz string, password string) {
	//manager := &ethereum.WalletManager{}

	//err := manager.CreateWalletFlow()
	_, path, err := ethereum.CreateNewWallet(aliaz, password)
	if err != nil {
		fmt.Println("create wallet failed, err=", err)
		return
	}

	fmt.Println("wallet path:", path)
}

func TestBatchCreateAddr() {
	manager := &ethereum.WalletManager{}

	err := manager.CreateAddressFlow()
	if err != nil {
		fmt.Println("CreateAddressFlow faile, err=%", err)
	}

	//ethereum.GetWalletList()
}

func TestBitInt() {
	i := new(big.Int)
	_, success := i.SetString("ff", 16)
	if success {
		fmt.Println("i:", i.String())
	}

	//fmt.Println("false, i:", i.String())
}
