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
	_, err := ethereum.PrepareForBlockScanTest(fromAddrs, passwords)
	if err != nil {
		fmt.Println("prepare for test failed, err=", err)
		return err
	}
	return nil
}

func TestDbInf() error {
	wallets, err := ethereum.GetWalletList()
	if err != nil {
		fmt.Println("get Wallet list failed, err=", err)
		return err
	}

	if len(wallets) == 0 {
		fmt.Println("no wallet found.")
		return err
	}
	wallets[len(wallets)-1].DumpWalletDB()
	ethereum.DumpBlockScanDb()
	return nil
}

func TestBlockScanWhenFork() error {
	//ethereum.OpenDB(ethereum.)
	db, err := ethereum.OpenDB("/Users/peter/workspace/bitcoin/wallet/src/github.com/blocktree/OpenWallet/test/data/eth/db", ethereum.BLOCK_CHAIN_DB)
	if err != nil {
		fmt.Println("open eth block scan db failed, err=", err)
		return err
	}

	//手动修改block的hash,
	blocknums := []string{
		"0x2a19f",
		"0x2a19e",
		"0x2a19d",
	}

	for i, _ := range blocknums {
		var theBlocks []ethereum.BlockHeader
		err = db.Find("BlockNumber", blocknums[i], &theBlocks)
		if err != nil {
			fmt.Println("find block bumber failed, err=", err)
			return err
		}

		for j, _ := range theBlocks {
			theBlocks[j].BlockHash = "123456"
			err = db.Update(&theBlocks[j])
			if err != nil {
				fmt.Println("update block bumber failed, err=", err)
				return err
			}
		}
	}

	db.Close()

	manager := &ethereum.WalletManager{}
	scanner := ethereum.NewETHBlockScanner(manager)
	wallets, err := ethereum.GetWalletList()
	if err != nil {
		fmt.Println("get Wallet list failed, err=", err)
		return err
	}

	if len(wallets) == 0 {
		fmt.Println("no wallet found.")
		return err
	}

	w := wallets[len(wallets)-1]
	err = scanner.AddWallet(w.WalletID, w)
	if err != nil {
		fmt.Println("scanner add wallet failed, err=", err)
		return err
	}

	scanner.ScanBlock()
	fmt.Println("after scan block, show db following:")
	w.DumpWalletDB()
	ethereum.DumpBlockScanDb()
	return nil
}

func TestBlockScan() error {
	fromAddrs := make([]string, 0, 2)
	passwords := make([]string, 0, 2)
	fromAddrs = append(fromAddrs, "0x50068fd632c1a6e6c5bd407b4ccf8861a589e776")
	passwords = append(passwords, "123456")
	fromAddrs = append(fromAddrs, "0x2a63b2203955b84fefe52baca3881b3614991b34")
	passwords = append(passwords, "123456")
	beginBlockNum, err := ethereum.PrepareForBlockScanTest(fromAddrs, passwords)
	if err != nil {
		fmt.Println("PrepareForBlockScanTest failed, err=", err)
		return err
	}

	manager := &ethereum.WalletManager{}
	scanner := ethereum.NewETHBlockScanner(manager)
	wallets, err := ethereum.GetWalletList()
	if err != nil {
		fmt.Println("get Wallet list failed, err=", err)
		return err
	}

	if len(wallets) == 0 {
		fmt.Println("no wallet found.")
		return err
	}

	w := wallets[len(wallets)-1]
	err = scanner.AddWallet(w.WalletID, w)
	if err != nil {
		fmt.Println("scanner add wallet failed, err=", err)
		return err
	}

	w.ClearAllTransactions()

	ethereum.ClearBlockScanDb()
	scanner.SetLocalBlock(beginBlockNum)
	scanner.ScanBlock()
	fmt.Println("after scan block, show db following:")
	w.DumpWalletDB()
	ethereum.DumpBlockScanDb()
	return nil
}
