package tech

import (
	//"log"

	"encoding/json"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

func createTransaction(walletID, accountID, to string) (*openwallet.RawTransaction, error) {

	tm.Init()
	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
		return nil, err
	}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, "0.0008", to, "", "", nil)
	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func createErc20TokenTransaction(walletID, accountID, to string) (*openwallet.RawTransaction, error) {
	tm.Init()
	err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	if err != nil {
		log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
		return nil, err
	}

	rawTx, err := tm.CreateErc20TokenTransaction(testApp, walletID, accountID, "1", to, "", "",
		"0x8847E5F841458ace82dbb0692C97115799fe28d3", "peterToken", "PTN", 1)
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
	tm.Init()
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
	tm.Init()
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
	tm.Init()
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

func TestWalletManager_SubmitTokenTransaction() {
	//0x2a7A2CcF5d899bB4cA4D7C381161B75a6A3778f1
	walletID := "W6EZ35wMPeYG7QJjVTpU6heCE4AxmkVzJd"
	accountID := "KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8"
	to := "0x584a9Ed7f95Cd04337df791Fac32bED88E13b77a"
	tm.Init()
	rawTx, err := createErc20TokenTransaction(walletID, accountID, to)
	if err != nil {
		return
	}

	_, err = tm.SignTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return
	}

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
}
func TestGetRawTransactionFeeRate() {
	manager, _ := GetEthWalletManager()
	//decoder := ethereum.EthTransactionDecoder{}
	decoder := manager.TxDecoder
	gasprice, gaslimit, err := decoder.GetRawTransactionFeeRate()
	if err != nil {
		log.Errorf("GetRawTransactionFeeRate failed, err=%v", err)
		return
	}
	log.Debugf("gasprice:%v gaslimit:%v", gasprice, gaslimit)
}

func TestWalletManager_SubmitTransaction() {

	//walletID := "W9cRnfgyZ7T4imjbQuiafz6Ca5aUf8qJRJ"
	//accountID := "4mNzv15wSPeUgqfw2Y4UieRJzUoJJMS9DM1L136gxFMZ"
	//to := "0xE1B74B188284A4323a1Cb95B130B00445628113e"
	//	tm.Init()
	walletID := "W6EZ35wMPeYG7QJjVTpU6heCE4AxmkVzJd"
	accountID := "KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8"
	to := "0xd35f9Ea14D063af9B3567064FAB567275b09f03D" //"0x584a9Ed7f95Cd04337df791Fac32bED88E13b77a"

	rawTx, err := createTransaction(walletID, accountID, to)
	if err != nil {
		return
	}

	//str, _ := json.MarshalIndent(rawTx, "", " ")
	//log.Info("create rawTx:", string(str))

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

	//str, _ = json.MarshalIndent(rawTx, "", " ")
	//log.Info("rawTx:", string(str))
}
