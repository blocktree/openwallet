package tech

import (
	"fmt"
	"log"
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
		fmt.Println("CreateAddressFlow failed, err=", err)
	}

	//ethereum.GetWalletList()
}

func TestBitInt() {
	i := new(big.Int)
	_, success := i.SetString("ff", 16)
	if success {
		fmt.Println("i:", i.String())
	}
}

func TestTransferFlow() {
	manager := &ethereum.WalletManager{}

	err := manager.TransferFlow()
	if err != nil {
		log.Fatal("transfer flow failed, err = ", err)
	}
}

func TestSummaryFlow() {
	manager := &ethereum.WalletManager{}

	err := manager.SummaryFollow()
	if err != nil {
		log.Fatal("summary flow failed, err = ", err)
	}
}

func TestBackupWallet() {
	manager := &ethereum.WalletManager{}

	err := manager.BackupWalletFlow()
	if err != nil {
		log.Fatal("backup wallet flow failed, err = ", err)
	}
}

func TestRestoreWallet() {
	manager := &ethereum.WalletManager{}

	err := manager.RestoreWalletFlow()
	if err != nil {
		log.Fatal("restore wallet flow failed, err = ", err)
	}
}
