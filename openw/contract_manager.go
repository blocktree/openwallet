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

package openw

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"time"
)

// CreateSmartContractTransaction
func (wm *WalletManager) CallSmartContractABI(appID, walletID, accountID string, contract *openwallet.SmartContract, abiParam []string) (*openwallet.SmartContractCallResult, error) {

	var (
		coin openwallet.Coin
	)

	if contract == nil {
		return nil, fmt.Errorf("contract is nil")
	}

	wrapper, err := wm.NewWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, err
	}

	assetsMgr, err := GetAssetsAdapter(account.Symbol)
	if err != nil {
		return nil, err
	}
	//fmt.Println("contract:", contract)

	coin = openwallet.Coin{
		Symbol:     account.Symbol,
		ContractID: contract.ContractID,
		IsContract: true,
		Contract:   *contract,
	}

	rawTx := openwallet.SmartContractRawTransaction{
		Coin:     coin,
		Account:  account,
		ABIParam: abiParam,
	}

	scDecoder := assetsMgr.GetSmartContractDecoder()
	if scDecoder == nil {
		return nil, fmt.Errorf("[%s] is not support smart contract transaction. ", account.Symbol)
	}

	result, callErr := scDecoder.CallSmartContractABI(wrapper, &rawTx)
	if callErr != nil {
		return nil, callErr
	}

	return result, nil
}

// CreateSmartContractTransaction
func (wm *WalletManager) CreateSmartContractTransaction(appID, walletID, accountID, amount, feeRate string, contract *openwallet.SmartContract, abiParam []string) (*openwallet.SmartContractRawTransaction, *openwallet.Error) {

	var (
		coin openwallet.Coin
	)

	if contract == nil {
		return nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, "contract is nil")
	}

	wrapper, err := wm.NewWalletWrapper(appID, "")
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}

	assetsMgr, err := GetAssetsAdapter(account.Symbol)
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}
	//fmt.Println("contract:", contract)

	coin = openwallet.Coin{
		Symbol:     account.Symbol,
		ContractID: contract.ContractID,
		IsContract: true,
		Contract:   *contract,
	}

	rawTx := openwallet.SmartContractRawTransaction{
		Coin:     coin,
		Account:  account,
		FeeRate:  feeRate,
		Value:    amount,
		ABIParam: abiParam,
	}

	scDecoder := assetsMgr.GetSmartContractDecoder()
	if scDecoder == nil {
		return nil, openwallet.Errorf(openwallet.ErrSystemException, "[%s] is not support smart contract transaction. ", account.Symbol)
	}

	createErr := scDecoder.CreateSmartContractRawTransaction(wrapper, &rawTx)
	if createErr != nil {
		return nil, createErr
	}

	log.Debug("transaction has been created successfully")

	return &rawTx, nil
}

// SubmitSmartContractTransaction
func (wm *WalletManager) SubmitSmartContractTransaction(appID, walletID, accountID string, rawTx *openwallet.SmartContractRawTransaction) (*openwallet.SmartContractReceipt, *openwallet.Error) {

	wrapper, err := wm.NewWalletWrapper(appID, "")
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}

	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}

	assetsMgr, err := GetAssetsAdapter(account.Symbol)
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}

	scdecoder := assetsMgr.GetSmartContractDecoder()
	if scdecoder == nil {
		return nil, openwallet.Errorf(openwallet.ErrSystemException, "[%s] is not support smart contract transaction. ", account.Symbol)
	}

	tx, submitErr := scdecoder.SubmitSmartContractRawTransaction(wrapper, rawTx)
	if submitErr != nil {
		return nil, submitErr
	}

	log.Debug("smart contract transaction has been submitted successfully")

	//log.Info("Save new transaction data successfully")
	//db, err := wrapper.OpenStormDB()
	//if err != nil {
	//	return tx, nil
	//}
	//defer wrapper.CloseDB()
	//
	////保存账户相关的记录
	//err = db.Save(tx)
	//if err != nil {
	//	return tx, nil
	//}

	return tx, nil
	//return perfectTx, nil
}

// SignSmartContractTransaction
func (wm *WalletManager) SignSmartContractTransaction(appID, walletID, accountID, password string, rawTx *openwallet.SmartContractRawTransaction) (*openwallet.SmartContractRawTransaction, *openwallet.Error) {

	account, err := wm.GetAssetsAccountInfo(appID, "", accountID)
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}

	wrapper, err := wm.NewWalletWrapper(appID, account.WalletID)
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}

	//解锁钱包
	err = wrapper.UnlockWallet(password, 5*time.Second)
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}

	key, err := wrapper.HDKey()
	if err != nil {
		return nil, openwallet.ConvertError(err)
	}

	for accountID, keySignatures := range rawTx.Signatures {
		//log.Infof("accountID: %s", accountID)
		if keySignatures != nil {
			for _, keySignature := range keySignatures {

				childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
				keyBytes, err := childKey.GetPrivateKeyBytes()
				if err != nil {
					return nil, openwallet.ConvertError(err)
				}
				//log.Debug("privateKey:", hex.EncodeToString(keyBytes))

				//privateKeys = append(privateKeys, keyBytes)
				txHash, err := hex.DecodeString(keySignature.Message)
				//transHash = append(transHash, txHash)

				//log.Infof("sign hash: %s", txHash)

				//签名交易
				/////////交易单哈希签名

				//signature, err := signatureSet.SignTxHash(rawTx.Coin.Symbol, txHash, keyBytes, keySignature.EccType)
				//if err != nil {
				//	return fmt.Errorf("transaction hash sign failed, unexpected error: %v", err)
				//}

				signature, v, sigErr := owcrypt.Signature(keyBytes, nil, txHash, keySignature.EccType)
				if sigErr != owcrypt.SUCCESS {
					return nil, openwallet.Errorf(openwallet.ErrSystemException,"transaction hash sign failed")
				}

				if keySignature.RSV {
					signature = append(signature, v)
				}

				//log.Debug("Signature:", txHash)

				keySignature.Signature = hex.EncodeToString(signature)
			}
		}
		rawTx.Signatures[accountID] = keySignatures
	}

	log.Debug("transaction has been signed successfully")

	return rawTx, nil
}