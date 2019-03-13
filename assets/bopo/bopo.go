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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/console"
	"github.com/blocktree/openwallet/timer"
	"github.com/shopspring/decimal"
)

func (w *WalletManager) InitConfigFlow() error {
	return errors.New("Writing1!")
}

func (w *WalletManager) ShowConfig() error {
	return errors.New("Writing2!")
}

func (w *WalletManager) GetWalletList() error {
	// Load config
	if err := w.loadConfig(); err != nil {
		return err
	}

	wallets, err := w.getWalletList()
	if err != nil {
		return err
	}

	w.printWalletList(wallets)
	return nil
}

func (w *WalletManager) CreateWalletFlow() error {
	// Load config
	if err := w.loadConfig(); err != nil {
		return err
	}

	name, err := console.InputText("Wallet name: ", true)
	if err != nil {
		return err
	}

	if wallet, err := w.createWallet(name); err != nil {
		return err
	} else {
		w.printWalletList([]*Wallet{wallet})
	}

	return nil
}

func (w *WalletManager) TransferFlow() error {
	var (
		wid     string // Wallet Alias
		toaddr  string // Wallet Address
		amount  string // Amount
		message string // Message to transfer (Uniform, and will be used to search as a global key)
		err     error
	)

	// Load config
	if err := w.loadConfig(); err != nil {
		return err
	}

	// Show all wallet addr
	if err := w.GetWalletList(); err != nil {
		return err
	}

	// Wallet ID
	for i := 0; i < 3; i++ {
		wid, err = console.InputText("Use wallet (By: alias): ", true)
		if err != nil {
			return err
		}

		// Check wid
		if _, err := w.getWalletInfo(wid); err != nil {
			fmt.Println(err)
		} else {
			break
		}

		// Stop after 3 times to check
		if i == 2 {
			return nil
		}
	}

	// To addr
	for i := 0; i < 3; i++ {
		toaddr, err = console.InputText("To address: ", true)
		if err != nil {
			return err
		}

		// Check addr
		if err := w.verifyAddr(toaddr); err != nil {
			fmt.Println(err)
		} else {
			break
		}

		// Stop after 3 times to check
		if i == 2 {
			return nil
		}
	}

	// Amount
	for i := 0; i < 3; i++ {
		amount, err = console.InputText("Amount(Unit: coin, 1 coin = 10^8 pais): ", true)
		if err != nil {
			fmt.Println(err)
			//return err
		}
		if cc, err := decimal.NewFromString(amount); err != nil {
			fmt.Println(err)
			// return err
		} else {
			amount = cc.Mul(w.config.coinDecimal).String()
			break
		}

		// Stop after 3 times to check
		if i == 2 {
			return nil
		}
	}

	// Message
	message = time.Now().UTC().Format(time.RFC850)

	if cfi, err := console.InputText(fmt.Sprintf("To addr[%s] with amount[%s] from alias[%s]'s account(yes/no)? ", toaddr, amount, wid), true); err != nil || cfi != "yes" {
		return err
	}

	fmt.Println("Transfer......")
	if wallet, err := w.toTransfer(wid, toaddr, amount, message); err != nil {
		return err
	} else {
		time.Sleep(12 * time.Second)
		// printWalletList([]*Wallet{wallet, &Wallet{Addr: toaddr}})
		w.printWalletList([]*Wallet{wallet})
	}

	return nil
}

func (w *WalletManager) BackupWalletFlow() error {

	// Load config
	if err := w.loadConfig(); err != nil {
		return err
	}

	if err := w.backupWalletData(); err != nil {
		return err
	}

	return nil
}

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (w *WalletManager) RestoreWalletFlow() error {

	var (
		datFile string // The filepath of wallet.dat which will be restored from local

		err error
	)

	// Load config
	if err := w.loadConfig(); err != nil {
		return err
	}

	// List all backup files
	files, err := FilePathWalkDir("./data")
	if err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for i, f := range files {
		fmt.Printf("%d>\t %v \n", i+1, filepath.Join(dir, f))
	}

	// Input dataFile of wallet.dat in Loop
	for i := 0; i < 3; i++ {

		datFile, err = console.InputText("Enter backup wallet.dat file path: ", true)
		if err != nil {
			fmt.Println(err)
		}
		if _, err := os.Stat(datFile); os.IsNotExist(err) {
			fmt.Println("No such file!")
		} else {
			break
		}

		// Stop after 3 times to check
		if i == 2 {
			return nil
		}

	}

	// To restore
	if err := w.restoreWalletData(datFile); err != nil {
		return err
	}

	return nil
}

func (w *WalletManager) CreateAddressFlow() error {
	return errors.New("A wallet only have one address, it's same thing in BOPO!")
}

func (w *WalletManager) SummaryFollow() error {

	fmt.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

	// Load config
	if err := w.loadConfig(); err != nil {
		return err
	}

	// Check summary addr
	if len(w.config.sumAddress) == 0 {
		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", Symbol))
	}
	if err := w.verifyAddr(w.config.sumAddress); err != nil {
		fmt.Println("Summary address invalid!")
		return err
	}

	// if w, err := w.getWalletInfo2(w.config.sumAddress); err != nil {
	// 	log.Println(err)
	// 	return err
	// } else {
	// 	fmt.Println("The summary address info: ")
	// 	w.printWalletList([]*Wallet{w})

	// 	// Confirm summary wallet
	// 	if cfi, err := console.InputText(fmt.Sprintf("Confirm  wid[%s](addr[%s]) to summary (yes/no)? ", w.WalletID, w.Addr), true); err != nil || cfi != "yes" {
	// 		return err
	// 	}
	// }

	//	// List all wallets that have balance to summary (without summaryAddr)
	//	wallets, err := getWalletList()
	//	if err != nil {
	//		return err
	//	}
	//	tmp := wallets[:0]
	//	for _, w := range wallets {
	//		// fmt.Printf("Summary: %d = %+v\t %+v\t %+v\t\n", 1, w.Alias, w.Balance, w.Balance == "")
	//		if w.Balance != "" && w.Addr != sumAddress {
	//			tmp = append(tmp, w)
	//		}
	//	}
	//	wallets = tmp
	//	fmt.Println("\nFollows will be summary")
	//	printWalletList(wallets)

	fmt.Printf("The timer for summary has started. Execute by every %v seconds.\n", w.config.cycleSeconds.Seconds())

	// Start timer
	sumTimer := timer.NewTask(w.config.cycleSeconds, w.SummaryWallets)
	sumTimer.Start()

	return nil
}
