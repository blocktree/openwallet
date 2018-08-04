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

package walletnode

import (
	"github.com/blocktree/OpenWallet/common/file"
	// "github.com/pkg/errors"
	"path/filepath"
	s "strings"
)

// Init and create <Symbol>.ini file automatically
func initConfig(symbol string) error {
	configFilePath, _ := filepath.Abs("conf")
	configFileName := s.ToUpper(symbol) + ".ini"

	absFile := filepath.Join(configFilePath, configFileName)
	if !file.Exists(absFile) {
		file.MkdirAll(configFilePath)
		file.WriteFile(absFile, []byte(defaultConfig), false)
	}
	return nil
}
