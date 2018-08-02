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
	// "fmt"
	// "github.com/blocktree/OpenWallet/common"
	// "github.com/blocktree/OpenWallet/console"
	// "github.com/blocktree/OpenWallet/logger"
	// "github.com/blocktree/OpenWallet/timer"
	// "github.com/shopspring/decimal"
	// "log"
	// "path/filepath"
	// "strings"
)

type WalletManager struct{}

func (w *WalletManager) InitConfigFlow() error {
	return errors.New("Writing!")
}

func (w *WalletManager) ShowConfig() error {
	return errors.New("Writing!")
}

func (w *WalletManager) CreateWalletFlow() error {

	if _, err := createNewWallet("c4"); err != nil {
		return err
	}
	return nil
}

func (w *WalletManager) CreateAddressFlow() error {
	return errors.New("Writing!")
}

func (w *WalletManager) SummaryFollow() error {
	return errors.New("Writing!")
}

func (w *WalletManager) BackupWalletFlow() error {
	return errors.New("Writing!")
}

func (w *WalletManager) GetWalletList() error {

	list, err := getWalletList()
	if err != nil {
		return err
	}

	printWalletList(list)
	return nil
}

func (w *WalletManager) TransferFlow() error {
	return errors.New("Writing!")
}

func (w *WalletManager) RestoreWalletFlow() error {
	return errors.New("Writing!")
}
