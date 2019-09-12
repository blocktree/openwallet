/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openw

import (
	"fmt"
	"time"

	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/shopspring/decimal"
)

func (wm *WalletManager) CreateErc20TokenTransaction(appID, walletID, accountID, amount, address, feeRate, memo,
	contractAddr, tokenName, tokenSymbol string, tokenDecimal uint64) (*openwallet.RawTransaction, error) {
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

	rawTx := openwallet.RawTransaction{
		Coin: openwallet.Coin{
			Symbol:     account.Symbol,
			ContractID: "",
			IsContract: true,
			Contract: openwallet.SmartContract{
				ContractID: "",
				Address:    contractAddr,
				Name:       tokenName,
				Symbol:     account.Symbol,
				Token:      tokenSymbol,
				Decimals:   tokenDecimal,
			},
		},
		Account:  account,
		FeeRate:  feeRate,
		To:       map[string]string{address: amount},
		Required: 1,
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}
	err = txdecoder.CreateRawTransaction(wrapper, &rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been created successfully")

	return &rawTx, nil
}

func (wm *WalletManager) CreateQrc20TokenTransaction(appID, walletID, accountID, sendAmount, toAddress, feeRate, memo,
	contractAddr, tokenName, tokenSymbol string, tokenDecimal uint64) (*openwallet.RawTransaction, error) {
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

	rawTx := openwallet.RawTransaction{
		Coin: openwallet.Coin{
			Symbol:     account.Symbol,
			ContractID: "",
			IsContract: true,
			Contract: openwallet.SmartContract{
				ContractID: "",
				Address:    contractAddr,
				Name:       tokenName,
				Symbol:     account.Symbol,
				Token:      tokenSymbol,
				Decimals:   tokenDecimal,
			},
		},
		Account: account,
		//gasPrice
		FeeRate:  feeRate,
		To:       map[string]string{toAddress: sendAmount},
		Required: 1,
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}
	err = txdecoder.CreateRawTransaction(wrapper, &rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("Qrc20Token transaction has been created successfully")

	return &rawTx, nil
}

// CreateTransaction
func (wm *WalletManager) CreateTransaction(appID, walletID, accountID, amount, address, feeRate, memo string, contract *openwallet.SmartContract) (*openwallet.RawTransaction, error) {

	var (
		coin openwallet.Coin
	)

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
	fmt.Println("contract:", contract)
	if contract != nil {
		coin = openwallet.Coin{
			Symbol:     account.Symbol,
			ContractID: contract.ContractID,
			IsContract: true,
			Contract:   *contract,
		}
	} else {
		coin = openwallet.Coin{
			Symbol:     account.Symbol,
			ContractID: "",
			IsContract: false,
		}
	}

	rawTx := openwallet.RawTransaction{
		Coin:     coin,
		Account:  account,
		FeeRate:  feeRate,
		To:       map[string]string{address: amount},
		Required: 1,
	}

	if len(memo) > 0 {
		rawTx.SetExtParam("memo", memo)
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	err = txdecoder.CreateRawTransaction(wrapper, &rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been created successfully")

	return &rawTx, nil
}

// SignTransaction
func (wm *WalletManager) SignTransaction(appID, walletID, accountID, password string, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	account, err := wm.GetAssetsAccountInfo(appID, "", accountID)
	if err != nil {
		return nil, err
	}

	wrapper, err := wm.NewWalletWrapper(appID, account.WalletID)
	if err != nil {
		return nil, err
	}

	assetsMgr, err := GetAssetsAdapter(account.Symbol)
	if err != nil {
		return nil, err
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	//解锁钱包
	err = wrapper.UnlockWallet(password, 5*time.Second)
	if err != nil {
		return nil, err
	}

	err = txdecoder.SignRawTransaction(wrapper, rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been signed successfully")

	return rawTx, nil
}

// VerifyTransaction
func (wm *WalletManager) VerifyTransaction(appID, walletID, accountID string, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

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

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	err = txdecoder.VerifyRawTransaction(wrapper, rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been validated successfully")

	return rawTx, nil
}

// SubmitTransaction
func (wm *WalletManager) SubmitTransaction(appID, walletID, accountID string, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

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

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	tx, err := txdecoder.SubmitRawTransaction(wrapper, rawTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been submitted successfully")

	log.Info("Save new transaction data successfully")
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return tx, nil
	}
	defer wrapper.CloseDB()

	//保存账户相关的记录
	err = db.Save(tx)
	if err != nil {
		return tx, nil
	}

	return tx, nil
	//return perfectTx, nil
}

//GetAssetsAccountBalance 获取账户余额
func (wm *WalletManager) GetAssetsAccountBalance(appID, walletID, accountID string) (*openwallet.Balance, error) {

	var (
		addressMap  = make(map[string]*openwallet.Address)
		searchAddrs = make([]string, 0)
	)

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

	//提取交易单
	scanner := assetsMgr.GetBlockScanner()
	if scanner == nil {
		return nil, fmt.Errorf("[%s] not support block scan", account.Symbol)
	}

	accountBalanceDec := decimal.New(0, 0)
	balances := make([]*openwallet.Balance, 0)
	//地址模型
	if assetsMgr.BalanceModelType() == openwallet.BalanceModelTypeAddress {

		addresses, addrErr := wrapper.GetAddressList(0, -1, "AccountID", accountID)
		if addrErr != nil {
			return nil, addrErr
		}

		for _, address := range addresses {
			searchAddrs = append(searchAddrs, address.Address)
			addressMap[address.Address] = address
		}

		balances, err = scanner.GetBalanceByAddress(searchAddrs...)
		if err != nil {
			return nil, err
		}

	} else if assetsMgr.BalanceModelType() == openwallet.BalanceModelTypeAccount { //账户模型
		balances, err = scanner.GetBalanceByAddress(account.Alias)
		if err != nil {
			return nil, err
		}

	}

	for _, b := range balances {
		//log.Debug("b.Balance:", b.Balance)
		addrBalance, _ := decimal.NewFromString(b.Balance)
		accountBalanceDec = accountBalanceDec.Add(addrBalance)
	}

	accountBalance := openwallet.Balance{
		Symbol:    account.Symbol,
		AccountID: accountID,
		Address:   "",
		Balance:   accountBalanceDec.StringFixed(assetsMgr.Decimal()),
	}

	return &accountBalance, nil
}

//GetAssetsAccountTokenBalance 获取账户Token余额
func (wm *WalletManager) GetAssetsAccountTokenBalance(appID, walletID, accountID string, contract openwallet.SmartContract) (*openwallet.TokenBalance, error) {

	var (
		addressMap  = make(map[string]*openwallet.Address)
		searchAddrs = make([]string, 0)
	)

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

	//提取交易单
	smartContractDecoder := assetsMgr.GetSmartContractDecoder()
	if smartContractDecoder == nil {
		return nil, fmt.Errorf("[%s] not support smart contract", account.Symbol)
	}

	accountBalanceDec := decimal.New(0, 0)
	balances := make([]*openwallet.TokenBalance, 0)
	//地址模型
	if assetsMgr.BalanceModelType() == openwallet.BalanceModelTypeAddress {

		addresses, addrErr := wrapper.GetAddressList(0, -1, "AccountID", accountID)
		if addrErr != nil {
			return nil, addrErr
		}

		for _, address := range addresses {
			searchAddrs = append(searchAddrs, address.Address)
			addressMap[address.Address] = address
		}

		balances, err = smartContractDecoder.GetTokenBalanceByAddress(contract, searchAddrs...)
		if err != nil {
			return nil, err
		}
	} else if assetsMgr.BalanceModelType() == openwallet.BalanceModelTypeAccount {
		balances, err = smartContractDecoder.GetTokenBalanceByAddress(contract, account.Alias)
		if err != nil {
			return nil, err
		}
	}

	for _, b := range balances {
		//log.Debug("b.Balance:", b.Balance)
		addrBalance, _ := decimal.NewFromString(b.Balance.Balance)
		accountBalanceDec = accountBalanceDec.Add(addrBalance)
	}

	accountBalance := openwallet.Balance{
		Symbol:    account.Symbol,
		AccountID: accountID,
		Address:   "",
		Balance:   accountBalanceDec.StringFixed(int32(contract.Decimals)),
	}

	accountTokenBalance := openwallet.TokenBalance{
		Contract: &contract,
		Balance:  &accountBalance,
	}

	return &accountTokenBalance, nil
}

//GetTransactions
func (wm *WalletManager) GetTransactions(appID string, offset, limit int, cols ...interface{}) ([]*openwallet.Transaction, error) {

	wrapper, err := wm.NewWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	txWrapper := NewTransactionWrapper(wrapper)
	trx, err := txWrapper.GetTransactions(offset, limit, cols...)
	if err != nil {
		return nil, err
	}

	return trx, nil
}

//GetTransactionByWxID 通过WxID获取交易单
func (wm *WalletManager) GetTransactionByWxID(appID, wxID string) (*openwallet.Transaction, error) {

	wrapper, err := wm.NewWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	txWrapper := NewTransactionWrapper(wrapper)
	trx, err := txWrapper.GetTransactions(0, -1, "WxID", wxID)
	if err != nil || len(trx) == 0 {
		return nil, err
	}

	return trx[0], nil
}

//GetTxUnspent
func (wm *WalletManager) GetTxUnspent(appID string, offset, limit int, cols ...interface{}) ([]*openwallet.TxOutPut, error) {

	wrapper, err := wm.NewWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	txWrapper := NewTransactionWrapper(wrapper)
	trx, err := txWrapper.GetTxOutputs(offset, limit, cols...)
	if err != nil {
		return nil, err
	}

	return trx, nil
}

//GetTxSpent
func (wm *WalletManager) GetTxSpent(appID string, offset, limit int, cols ...interface{}) ([]*openwallet.TxInput, error) {

	wrapper, err := wm.NewWalletWrapper(appID, "")
	if err != nil {
		return nil, err
	}

	txWrapper := NewTransactionWrapper(wrapper)
	trx, err := txWrapper.GetTxInputs(offset, limit, cols...)
	if err != nil {
		return nil, err
	}

	return trx, nil
}

//GetEstimateFeeRate 获取币种推荐手续费
func (wm *WalletManager) GetEstimateFeeRate(coin openwallet.Coin) (feeRate string, unit string, err error) {

	assetsMgr, err := GetAssetsAdapter(coin.Symbol)
	if err != nil {
		return "", "", err
	}

	txDecoder := assetsMgr.GetTransactionDecoder()
	if txDecoder == nil {
		return "", "", fmt.Errorf("[%s] is not support transaction. ", coin.Symbol)
	}

	return txDecoder.GetRawTransactionFeeRate()

}

// CreateSummaryTransaction
func (wm *WalletManager) CreateSummaryTransaction(
	appID, walletID, accountID, summaryAddress, minTransfer, retainedBalance, feeRate string,
	start, limit int,
	contract *openwallet.SmartContract) ([]*openwallet.RawTransaction, error) {

	var (
		coin openwallet.Coin
	)

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

	if contract != nil {
		coin = openwallet.Coin{
			Symbol:     account.Symbol,
			ContractID: contract.ContractID,
			IsContract: true,
			Contract:   *contract,
		}
	} else {
		coin = openwallet.Coin{
			Symbol:     account.Symbol,
			ContractID: "",
			IsContract: false,
		}
	}

	sumTx := openwallet.SummaryRawTransaction{
		Coin:              coin,
		Account:           account,
		FeeRate:           feeRate,
		SummaryAddress:    summaryAddress,
		MinTransfer:       minTransfer,
		RetainedBalance:   retainedBalance,
		AddressStartIndex: start,
		AddressLimit:      limit,
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	rawTxArray, err := txdecoder.CreateSummaryRawTransaction(wrapper, &sumTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been created successfully")

	return rawTxArray, nil
}

// CreateSummaryTransaction
func (wm *WalletManager) CreateSummaryRawTransactionWithError(
	appID, walletID, accountID, summaryAddress, minTransfer, retainedBalance, feeRate string,
	start, limit int,
	contract *openwallet.SmartContract,
	feeSupportAccount *openwallet.FeesSupportAccount,
) ([]*openwallet.RawTransactionWithError, error) {

	var (
		coin openwallet.Coin
	)

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

	if contract != nil {
		coin = openwallet.Coin{
			Symbol:     account.Symbol,
			ContractID: contract.ContractID,
			IsContract: true,
			Contract:   *contract,
		}
	} else {
		coin = openwallet.Coin{
			Symbol:     account.Symbol,
			ContractID: "",
			IsContract: false,
		}
	}

	sumTx := openwallet.SummaryRawTransaction{
		Coin:               coin,
		Account:            account,
		FeeRate:            feeRate,
		SummaryAddress:     summaryAddress,
		MinTransfer:        minTransfer,
		RetainedBalance:    retainedBalance,
		AddressStartIndex:  start,
		AddressLimit:       limit,
		FeesSupportAccount: feeSupportAccount,
	}

	txdecoder := assetsMgr.GetTransactionDecoder()
	if txdecoder == nil {
		return nil, fmt.Errorf("[%s] is not support transaction. ", account.Symbol)
	}

	rawTxArray, err := txdecoder.CreateSummaryRawTransactionWithError(wrapper, &sumTx)
	if err != nil {
		return nil, err
	}

	log.Debug("transaction has been created successfully")

	return rawTxArray, nil
}
