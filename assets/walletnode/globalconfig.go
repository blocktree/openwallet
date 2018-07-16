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
	"encoding/json"
	"fmt"
	bconfig "github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/pkg/errors"
	"path/filepath"
	s "strings"
)

// Load settings for global from local conf/<Symbol>.ini
func loadConfig(symbol string) error {
	var (
		c   bconfig.Configer
		err error
	)
	configFilePath, _ := filepath.Abs("conf")
	configFileName := s.ToUpper(symbol) + ".ini"

	absFile := filepath.Join(configFilePath, configFileName)
	c, err = bconfig.NewConfig("ini", absFile)
	if err != nil {
		return errors.New(fmt.Sprintf("Load Config Failed: %s", err))
	}

	if v, err := c.Bool("isTestNet"); err != nil {
		return errors.New(fmt.Sprintf("Load Config Failed: %s", err))
	} else {
		isTestNet = v
	}

	return nil
}

// Init <Symbol>.ini file automatically
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

// Update <Symbol>.ini file
func updateConfig(symbol string) error {
	if err := loadConfig(symbol); err != nil {
		return err
	}
	configFilePath, _ := filepath.Abs("conf")
	configFileName := s.ToUpper(symbol) + ".ini"
	absFile := filepath.Join(configFilePath, configFileName)
	fmt.Println(absFile)

	configMap := map[string]interface{}{
		"dockerAddr": dockerAddr,
		"dockerPort": dockerPort,
		"isTestNet":  isTestNet,
	}

	if bytes, err := json.Marshal(configMap); err != nil {
		return err
	} else {
		//实例化配置
		if c, err := bconfig.NewConfigData("json", bytes); err != nil {
			return err
		} else {
			//写入配置到文件
			fmt.Println(c)
			//   file.MkdirAll(configFilePath)
			//   absFile := filepath.Join(configFilePath, configFileName)
			//   err = c.SaveConfigFile(absFile)
			//   if err != nil {
			//   	return err
			//   }
		}
	}

	return nil
}
