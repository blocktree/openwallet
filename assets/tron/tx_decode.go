/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package tron

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv" // "sort"
	// "github.com/blocktree/OpenWallet/log"
	"strings"
	"time"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/shopspring/decimal"
	// "github.com/blocktree/OpenWallet/assets/qtum/btcLikeTxDriver"
	// "github.com/blocktree/OpenWallet/log"
	// "github.com/shopspring/decimal"
)

//TransactionDecoder for Interface TransactionDecode
type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

func CheckRawTransaction(rawTx *openwallet.RawTransaction) error {
	//账户模型原始账单只有一个To
	if len(rawTx.To) != 1 {
		return fmt.Errorf("noly one to address can be set!")
	}
	return nil
}

func InsertSignatureIntoRawTransaction(txHex string, signature string) (string, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		log.Errorf("invalid transaction hex data,err=", err)
		return "", fmt.Errorf("invalid transaction hex data!")
	}
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		log.Errorf("invalid transaction signature hex data,err=", err)
		return "", fmt.Errorf("invalid signature hex data!")
	}
	mergeTxBytes := append(txBytes, signatureBytes...)
	mergeTxHex := hex.EncodeToString(mergeTxBytes)
	return mergeTxHex, nil

}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
	)
	if rawTx.Coin.Symbol != Symbol {
		return errors.New("CreateRawTransaction: Symbol is not <TRX>")
	}

	if len(rawTx.To) == 0 {
		return fmt.Errorf("CreateRawTransaction: Receiver addresses is empty")
	}
	if rawTx.Account.AccountID == "" {
		return fmt.Errorf("CreateRawTransaction: AccountID is empty")
	}
	addressList, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)
	if err != nil {
		log.Errorf("get address list failed,err=", err)
		return err
	}
	if len(addressList) == 0 {
		return fmt.Errorf("[%s] account: %s has not addresses", decoder.wm.Symbol(), rawTx.Account.AccountID)
	}
	addressesBalanceList := make([]AddrBalance, 0, len(addressList))
	for i, addr := range addressList {
		balance, err := decoder.wm.Getbalance(addr.Address)
		if err != nil {
			log.Errorf("get balance failed,err=", err)
			return err
		}
		balance.Index = i
		addressesBalanceList = append(addressesBalanceList, *balance)
	}
	sort.Slice(addressesBalanceList, func(i int, j int) bool {
		return addressesBalanceList[i].TronBalance.Cmp(addressesBalanceList[j].TronBalance) >= 0
	})
	var amountStr, toAddress string
	for k, v := range rawTx.To {
		toAddress = k
		amountStr = v
		break
	}
	//计算账户的实际转账amount
	accountTotalSentAddresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID, "Address", toAddress)
	if findErr != nil || len(accountTotalSentAddresses) == 0 {
		amountDec, _ := decimal.NewFromString(amountStr)
		accountTotalSent = accountTotalSent.Add(amountDec)
	}
	txTo = []string{fmt.Sprintf("%s:%s", toAddress, amountStr)}
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		log.Errorf("conver amount from string  to float failed,err=", err)
		return err
	}
	signatureMap := make(map[string][]*openwallet.KeySignature)
	keySignList := make([]*openwallet.KeySignature, 0, 1)
	amountInt64 := int64(amountFloat * 1000000)
	amount := big.NewInt(amountInt64)
	count := big.NewInt(0)
	countList := []uint64{}
	for i, _ := range addressesBalanceList {
		if addressesBalanceList[i].TronBalance.Cmp(amount) < 0 {
			count.Add(count, addressesBalanceList[i].TronBalance)
			if count.Cmp(amount) >= 0 {
				countList = append(countList, addressesBalanceList[i].TronBalance.Sub(addressesBalanceList[i].TronBalance, count.Sub(count, amount)).Uint64())
				return fmt.Errorf("The Tron of account is enough,"+
					"but cannot be sent in just one transaction!\n"+
					"the amount can be sent in "+string(len(countList))+"times with amounts:\n"+strings.Replace(strings.Trim(fmt.Sprint(countList), "[]"), " ", ",", -1), err)
			} else {
				countList = append(countList, addressesBalanceList[i].TronBalance.Uint64())
			}
			continue
		} else {
			ownerAddress := &openwallet.Address{
				AccountID:   addressList[addressesBalanceList[i].Index].AccountID,
				Address:     addressList[addressesBalanceList[i].Index].Address,
				PublicKey:   addressList[addressesBalanceList[i].Index].PublicKey,
				Alias:       addressList[addressesBalanceList[i].Index].Alias,
				Tag:         addressList[addressesBalanceList[i].Index].Tag,
				Index:       addressList[addressesBalanceList[i].Index].Index,
				HDPath:      addressList[addressesBalanceList[i].Index].HDPath,
				WatchOnly:   addressList[addressesBalanceList[i].Index].WatchOnly,
				Symbol:      addressList[addressesBalanceList[i].Index].Symbol,
				Balance:     addressList[addressesBalanceList[i].Index].Balance,
				IsMemo:      addressList[addressesBalanceList[i].Index].IsMemo,
				Memo:        addressList[addressesBalanceList[i].Index].Memo,
				CreatedTime: addressList[addressesBalanceList[i].Index].CreatedTime,
			}
			keySignList = append(keySignList, &openwallet.KeySignature{
				Address: ownerAddress,
			})
			txFrom = []string{fmt.Sprintf("%s:%s", ownerAddress.Address, amountStr)}
			break
		}

	}
	//fmt.Println("from address:=", keySignList[0].Address.Address)
	if len(keySignList) != 1 {
		return fmt.Errorf("NO enough Tron to send")
	}

	//创建空交易单
	rawHex, err := decoder.wm.CreateTransactionRef(toAddress, keySignList[0].Address.Address, amountFloat)
	if err != nil {
		return err
	}

	txHashBytes, err := getTxHash1(rawHex)
	if err != nil {
		log.Errorf("get Tx hash failed,err=", err)
		return err
	}
	txHash := hex.EncodeToString(txHashBytes)
	keySignList[0].Nonce = ""
	keySignList[0].Message = txHash
	signatureMap[rawTx.Account.AccountID] = keySignList
	//rawTx.Signatures = make(map[string][]*openwallet.KeySignature, 0)

	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.Fees = "0"
	rawTx.FeeRate = "0"
	rawTx.RawHex = rawHex
	rawTx.Signatures = signatureMap
	rawTx.IsBuilt = true
	rawTx.TxTo = txTo
	rawTx.TxFrom = txFrom
	rawTx.TxAmount = accountTotalSent.StringFixed(decoder.wm.Decimal())
	return nil
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	key, err := wrapper.HDKey()
	if err != nil {
		log.Errorf("wrapper HDkey error,err=", err)
		return err
	}
	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]
	//fmt.Println("keySignatures:=", keySignatures)
	if keySignatures != nil {
		for _, keySignature := range keySignatures {
			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, 0xECC00000)
			if err != nil {
				log.Errorf("derived key with path failed,err=", err)
				return err
			}
			priKeyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				log.Errorf("get privatekey bytes failed,err=", err)
				return err
			}
			txHashBytes, err := getTxHash1(rawTx.RawHex)
			if err != nil {
				log.Errorf("get Tx hash failed,err=", err)
				return err
			}
			txHash := hex.EncodeToString(txHashBytes)
			priKey := hex.EncodeToString(priKeyBytes)
			signature, err := decoder.wm.SignTransactionRef(txHash, priKey)
			if err != nil {
				log.Errorf("sign Tx failed,err=", err)
				return err
			}
			keySignature.Signature = signature
		}
	}
	log.Info("Tx hash sign success!")
	//rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	//检测交易单基本字段
	err := CheckRawTransaction(rawTx)
	if err != nil {
		log.Errorf("verify Tx base field failed,err=", err)
		return err
	}
	if len(rawTx.Signatures) != 1 {
		return fmt.Errorf("the length of signature error")
	}
	sig, exist := rawTx.Signatures[rawTx.Account.AccountID]
	if !exist {
		return fmt.Errorf("wallet signature not found")
	}
	mergeTxHex, err := InsertSignatureIntoRawTransaction(rawTx.RawHex, sig[0].Signature)
	if err != nil {
		log.Errorf("merge empty transaction and signature failed,err=", err)
		return err
	}
	verifyRet := decoder.wm.ValidSignedTransactionRef(mergeTxHex)
	if verifyRet != nil {
		log.Error("Tx signature verify failed, err=", verifyRet)
		return fmt.Errorf("Tx signature verify failed, err=")
	} else {
		rawTx.IsCompleted = true
		//rawTx.RawHex = mergeTxHex
	}
	return nil
}

//SubmitRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}
	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}

	//********合并交易单********
	sig, exist := rawTx.Signatures[rawTx.Account.AccountID]
	if !exist {
		return nil, fmt.Errorf("wallet signature not found")
	}
	mergeTxHex, err := InsertSignatureIntoRawTransaction(rawTx.RawHex, sig[0].Signature)
	if err != nil {
		log.Errorf("merge empty transaction and signature failed,err=", err)
		return nil, err
	}
	rawTx.RawHex = mergeTxHex
	//********广播交易单********
	txid, err := decoder.wm.BroadcastTransaction(rawTx.RawHex)
	if err != nil {
		log.Errorf("submit transaction failed,err=", err)
		return nil, err
	}
	rawTx.TxID = txid
	rawTx.IsSubmit = true
	decimals := decoder.wm.Decimal()
	tx := openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       rawTx.Fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(&tx)
	return &tx, nil
}
