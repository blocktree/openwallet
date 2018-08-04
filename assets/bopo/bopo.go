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

package bopo

import (
	"errors"
	"github.com/blocktree/OpenWallet/console"
	"time"
)

type WalletManager struct{}

func (w *WalletManager) InitConfigFlow() error {
	return errors.New("Writing1!")
}

func (w *WalletManager) ShowConfig() error {
	return errors.New("Writing2!")
}

func (w *WalletManager) GetWalletList() error {
	// Load config
	if err := loadConfig(); err != nil {
		return err
	}

	wallets, err := getWalletList()
	if err != nil {
		return err
	}

	printWalletList(wallets)
	return nil
}

func (w *WalletManager) CreateWalletFlow() error {
	// Load config
	if err := loadConfig(); err != nil {
		return err
	}

	name, err := console.InputText("Wallet name: ", true)
	if err != nil {
		return err
	}

	if wallet, err := createWallet(name); err != nil {
		return err
	} else {
		printWalletList([]*Wallet{wallet})
	}

	return nil
}

func (w *WalletManager) TransferFlow() error {
	// Load config
	if err := loadConfig(); err != nil {
		return err
	}

	// Show all wallet addr
	if err := w.GetWalletList(); err != nil {
		return err
	}

	// Wallet ID
	wid, err := console.InputText("Use wallet (By: alias): ", true)
	if err != nil {
		return err
	}
	// To addr
	toaddr, err := console.InputText("To address: ", true)
	if err != nil {
		return err
	}
	// Amount
	amount, err := console.InputText("Amount(Unit: pais, 1 bopo = 10^8 pais): ", true)
	if err != nil {
		return err
	}
	message := time.Now().UTC().Format(time.RFC850)

	if wallet, err := toTransfer(wid, toaddr, amount, message); err != nil {
		return err
	} else {
		printWalletList([]*Wallet{wallet})
	}

	return nil
}

func (w *WalletManager) BackupWalletFlow() error {
	// Load config
	if err := loadConfig(); err != nil {
		return err
	}

	return errors.New("Writing!")
}

func (w *WalletManager) RestoreWalletFlow() error {
	// Load config
	if err := loadConfig(); err != nil {
		return err
	}

	return errors.New("Writing!")
}

func (w *WalletManager) CreateAddressFlow() error {
	return errors.New("Writing!")
}

func (w *WalletManager) SummaryFollow() error {
	return errors.New("Writing!")
}
