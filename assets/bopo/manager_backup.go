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
	"path/filepath"
	"strings"

	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/walletnode"
)

func (wm *WalletManager) backupWalletData() error {

	newBackupDir := filepath.Join(backupDir,
		strings.ToLower(Symbol)+"-"+common.TimeFormat("20060102150405"))
	file.MkdirAll(newBackupDir) // os.MkdirAll(backupDir, os.ModePerm)

	src := filepath.Join(walletDataPath, "wallet.dat")
	dst := filepath.Join(newBackupDir, "wallet.dat")

	if err := walletnode.CopyFromContainer(Symbol, src, dst); err != nil {
		return err
	}

	return nil
}

func (wm *WalletManager) restoreWalletData(datFile string) error {
	src := datFile
	dst := walletDataPath

	if err := walletnode.CopyToContainer(Symbol, src, dst); err != nil {
		return err
	}

	return nil
}
