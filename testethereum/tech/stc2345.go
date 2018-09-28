package tech

import (
	"encoding/json"
	"os"

	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/assets/ethereum"
	"github.com/blocktree/OpenWallet/assets/stc2345"
	"github.com/blocktree/OpenWallet/log"
)

func DumpWalletDB(dbPath string, dbfile string) {
	db, err := ethereum.OpenDB(dbPath, dbfile)
	if err != nil {
		log.Error("open db failed, err = ", err)
		return
	}
	defer db.Close()

	var wallets []ethereum.Wallet
	err = db.All(&wallets)
	if err != nil {
		log.Error("get address failed, err=", err)
		return
	}

	jsonStr, _ := json.MarshalIndent(wallets, "", " ")
	log.Debugf("wallets : %v", string(jsonStr))

	var addresses []ethereum.Address
	err = db.All(&addresses)
	if err != nil {
		log.Error("get address failed, err=", err)
		return
	}

	jsonStr, _ = json.MarshalIndent(addresses, "", " ")
	log.Debugf("wallets : %v", string(jsonStr))
}

func DumpEtc2345WalletDb() {
	DumpWalletDB("data/stc2345/db", "peter-W7isVa7z1JdiC9LanGSPApdsKZfWuACbwk.db")
}

func Getetc2345WalletManager() (*stc2345.WalletManager, error) {
	manager, ok := assets.Managers["stc2345"]
	if !ok {
		log.Error("cannot find the stc2345 manager.")
		os.Exit(-1)
	}
	return manager.(*stc2345.WalletManager), nil
}

func GetEthWalletManager() (*ethereum.WalletManager, error) {
	manager, ok := assets.Managers["eth"]
	if !ok {
		log.Error("cannot find the ethereum manager.")
		os.Exit(-1)
	}
	return manager.(*ethereum.WalletManager), nil
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

func TestTransferFlow2345() {
	manager, _ := Getetc2345WalletManager()
	err := manager.TransferFlow()
	if err != nil {
		log.Debugf("transfer flow failed, err = ", err)
	}
}

func TestSummaryFlow2345() {
	manager, _ := Getetc2345WalletManager()

	err := manager.SummaryFollow()
	if err != nil {
		log.Debugf("summary flow failed, err = ", err)
	}
}

func TestBackupWallet2345() {
	manager, _ := Getetc2345WalletManager() //manager := &ethereum.WalletManager{}

	err := manager.BackupWalletFlow()
	if err != nil {
		log.Debugf("backup wallet flow failed, err = ", err)
	}
}

func TestRestoreWallet2345() {
	manager, _ := Getetc2345WalletManager()

	err := manager.RestoreWalletFlow()
	if err != nil {
		log.Debugf("restore wallet flow failed, err = ", err)
	}
}
