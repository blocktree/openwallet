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
	"errors"

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
		if file.MkdirAll(configFilePath) != true {
			return errors.New("initConfig: MkdirAll failed!")
		}
		if file.WriteFile(absFile, []byte(WNConfig.defaultConfig), false) != true {
			return errors.New("initConfig: WriteFile failed!")
		}
	}
	return nil
}

// //读取配置
// absFile := filepath.Join(wc.configFilePath, wc.configFileName)
// if !file.Exists(absFile) {
// 	file.MkdirAll(wc.configFilePath)
// 	file.WriteFile(absFile, []byte(wc.defaultConfig), false)
// }
