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
	"github.com/blocktree/OpenWallet/common"
	"github.com/shopspring/decimal"
)

//GetWalletBalance 获取钱包余额
func (wm *WalletManager) GetWalletBalance(accountID string) string {

	balance := decimal.New(0, 0)

	return balance.StringFixed(8)
}

// //DumpWallet 导出钱包所有私钥文件
// func (wm *WalletManager) DumpWallet(filename string) error {

// 	request := []interface{}{
// 		filename,
// 	}

// 	_, err := wm.WalletClient.Call("dumpwallet", request)
// 	if err != nil {
// 		return err
// 	}

// 	return nil

// }

// //ImportWallet 导入钱包私钥文件
// func (wm *WalletManager) ImportWallet(filename string) error {

// 	request := []interface{}{
// 		filename,
// 	}

// 	_, err := wm.WalletClient.Call("importwallet", request)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

//SummaryWallets 执行汇总流程
func (wm *WalletManager) SummaryWallets() {

	wm.Log.Info("[Summary Wallet Start]------%s", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for wid, wallet := range wm.WalletsInSum {

		//统计钱包最新余额
		wb := wm.GetWalletBalance(wid)

		balance, _ := decimal.NewFromString(wb)
		//如果余额大于阀值，汇总的地址
		if balance.GreaterThan(wm.Config.Threshold) {

			wm.Log.Info("Summary account[%s]balance = %v ", wallet.WalletID, balance)
			wm.Log.Info("Summary account[%s]Start Send Transaction", wallet.WalletID)

			txID, err := wm.SendTransaction(wallet.WalletID, wm.Config.SumAddress, balance, wallet.Password, false)
			if err != nil {
				wm.Log.Info("Summary account[%s]unexpected error: %v", wallet.WalletID, err)
				continue
			} else {
				wm.Log.Info("Summary account[%s]successfully，Received Address[%s], TXID：%s", wallet.WalletID, wm.Config.SumAddress, txID)
			}
		} else {
			wm.Log.Info("Wallet Account[%s]-[%s]Current Balance: %v，below threshold: %v", wallet.Alias, wallet.WalletID, balance, wm.Config.Threshold)
		}
	}

	wm.Log.Info("[Summary Wallet end]------%s", common.TimeFormat("2006-01-02 15:04:05"))
}
