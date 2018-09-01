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

func TestConfigErcToken() {
	manager := &ethereum.WalletManager{}

	err := manager.ConfigERC20Token()
	if err != nil {
		log.Fatal("config erc20 token failed, err = ", err)
	}
}

func TestERC20TokenTransfer() {
	manager := &ethereum.WalletManager{}

	err := manager.ERC20TokenTransferFlow()
	if err != nil {
		log.Fatal("transfer erc20 token failed, err = ", err)
	}
}

func TestERC20TokenSummary() {
	manager := &ethereum.WalletManager{}

	err := manager.ERC20TokenSummaryFollow()
	if err != nil {
		log.Fatal("summary erc20 token failed, err = ", err)
	}
}

func PrepareTestForBlockScan() error {
	/*pending, queued, err := ethereum.EthGetTxpoolStatus()
	if err != nil {
		log.Fatal("get txpool status failed, err=", err)
		return
	}
	fmt.Println("pending number is ", pending, " queued number is ", queued)*/

	fromAddrs := make([]string, 0, 2)
	passwords := make([]string, 0, 2)
	fromAddrs = append(fromAddrs, "0x50068fd632c1a6e6c5bd407b4ccf8861a589e776")
	passwords = append(passwords, "123456")
	fromAddrs = append(fromAddrs, "0x2a63b2203955b84fefe52baca3881b3614991b34")
	passwords = append(passwords, "123456")
	err := ethereum.PrepareForBlockScanTest(fromAddrs, passwords)
	if err != nil {
		fmt.Println("prepare for test failed, err=", err)
		return err
	}
	return nil
}

func TestDbInf() error {
	wallet, err := ethereum.GetWalletList()
	if err != nil {
		fmt.Println("get Wallet list failed, err=", err)
		return err
	}

	if len(wallet) == 0 {
		fmt.Println("no wallet found.")
		return err
	}
	wallet[len(wallet)-1].DumpWalletDB()
	return nil
}
