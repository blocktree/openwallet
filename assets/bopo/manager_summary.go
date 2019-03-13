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

package bopo

import (
	"fmt"
	"time"

	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
)

//SummaryWallets 执行汇总流程
func (wm *WalletManager) SummaryWallets() {

	log.Std.Info("[Summary Wallet Start]------%s", common.TimeFormat("2006-01-02 15:04:05"))

	// //读取参与汇总的钱包
	// for wid, wallet := range wm.WalletsInSum {

	// 	//重新加载utxo
	// 	wm.RebuildWalletUnspent(wid)

	// 	//统计钱包最新余额
	// 	wb := wm.GetWalletBalance(wid)

	// 	balance, _ := decimal.NewFromString(wb)
	// 	//如果余额大于阀值，汇总的地址
	// 	if balance.GreaterThan(wm.Config.Threshold) {

	// 		log.Std.Info("Summary account[%s]balance = %v ", wallet.WalletID, balance)
	// 		log.Std.Info("Summary account[%s]Start Send Transaction", wallet.WalletID)

	// 		txID, err := wm.SendTransaction(wallet.WalletID, wm.Config.SumAddress, balance, wallet.Password, false)
	// 		if err != nil {
	// 			log.Std.Info("Summary account[%s]unexpected error: %v", wallet.WalletID, err)
	// 			continue
	// 		} else {
	// 			log.Std.Info("Summary account[%s]successfully，Received Address[%s], TXID：%s", wallet.WalletID, wm.Config.SumAddress, txID)
	// 		}
	// 	} else {
	// 		log.Std.Info("Wallet Account[%s]-[%s]Current Balance: %v，below threshold: %v", wallet.Alias, wallet.WalletID, balance, wm.Config.Threshold)
	// 	}
	// }

	// List all wallets that have balance to summary (without summaryAddr)
	wallets, err := wm.getWalletList()
	if err != nil {
		log.Info(err)
	}

	tmp := wallets[:0]
	for _, w := range wallets {
		if w.Balance != "" && w.Addr != wm.config.sumAddress {
			tmp = append(tmp, w)
		}
	}
	wallets = tmp
	fmt.Printf("There are %d wallets will be summary······", len(wallets))
	// if len(wallets) > 0 {
	// 	printWalletList(wallets)		// 加上后，会转账失败！原因未知？留作提醒
	// }

	// Summary
	for _, w := range wallets {
		message := time.Now().UTC().Format(time.RFC850)

		if _, err := wm.toTransfer(w.WalletID, wm.config.sumAddress, w.Balance, message); err != nil {
			log.Error(err)
		}
	}

	log.Std.Info("[Summary Wallet end]------%s", common.TimeFormat("2006-01-02 15:04:05"))

}

//AddWalletInSummary 添加汇总钱包账户
func (wm *WalletManager) AddWalletInSummary(wid string, wallet *openwallet.Wallet) {
	wm.walletsInSum[wid] = wallet
}
