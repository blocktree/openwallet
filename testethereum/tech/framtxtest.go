package tech

import (
	//"log"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

func createTransaction(walletID, accountID, to string) (*openwallet.RawTransaction, error) {

	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
		return nil, err
	}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, "0.2", to, "", "")
	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func TestWalletManager_CreateTransaction() {

	walletID := "W9Azzt5LAttoyaYufQHZKsbqkwmZbNPM95"
	accountID := "KgfVLAp8FNKE8UW8in2aCsK638RQAcSjSkMEdP4pne8VKrffjt"
	to := "mySLanFVyyUL6P2fAsbiwfTuBBh9xf3vhT"

	rawTx, err := createTransaction(walletID, accountID, to)

	if err != nil {
		return
	}

	log.Info("rawTx:", rawTx)

}
