package tech

import (
	//"log"

	"encoding/json"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

func createTransaction(walletID, accountID, to string) (*openwallet.RawTransaction, error) {

	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
		return nil, err
	}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, "5", to, "", "")
	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func TestWalletManager_CreateTransaction() {
	walletID := "W6EZ35wMPeYG7QJjVTpU6heCE4AxmkVzJd"
	accountID := "KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8"
	to := "0xdb9a569f7b80030956dc9686b89D5fF15922E175"

	rawTx, err := createTransaction(walletID, accountID, to)

	if err != nil {
		return
	}

	str, _ := json.MarshalIndent(rawTx, "", " ")
	log.Info("rawTx:", string(str))
}

func TestWalletManager_SignTransaction() {

	walletID := "W6EZ35wMPeYG7QJjVTpU6heCE4AxmkVzJd"
	accountID := "KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8"
	to := "0xdb9a569f7b80030956dc9686b89D5fF15922E175"

	rawTx, err := createTransaction(walletID, accountID, to)
	if err != nil {
		return
	}

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return
	}

	str, _ := json.MarshalIndent(rawTx, "", " ")
	log.Info("rawTx:", string(str))
}

func TestWalletManager_VerifyTransaction() {
	walletID := "W6EZ35wMPeYG7QJjVTpU6heCE4AxmkVzJd"
	accountID := "KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8"
	to := "0xdb9a569f7b80030956dc9686b89D5fF15922E175"

	rawTx, err := createTransaction(walletID, accountID, to)
	if err != nil {
		return
	}

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return
	}

	str, _ := json.MarshalIndent(rawTx, "", " ")
	log.Info("rawTx:", string(str))

	_, err = tm.VerifyTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return
	}

	str, _ = json.MarshalIndent(rawTx, "", " ")
	log.Info("rawTx:", string(str))

}

func TestWalletManager_SubmitTransaction() {

	walletID := "W6EZ35wMPeYG7QJjVTpU6heCE4AxmkVzJd"
	accountID := "KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8"
	to := "0x584a9Ed7f95Cd04337df791Fac32bED88E13b77a"

	rawTx, err := createTransaction(walletID, accountID, to)
	if err != nil {
		return
	}

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return
	}

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err = tm.VerifyTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return
	}

	str, _ := json.MarshalIndent(rawTx, "", " ")
	log.Info("rawTx:", string(str))

	_, err = tm.SubmitTransaction(testApp, walletID, accountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return
	}

	str, _ = json.MarshalIndent(rawTx, "", " ")
	log.Info("rawTx:", string(str))
}
