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

package merchant

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/common/file"
	"path/filepath"
)

var (

	//商户资料文件夹
	merchantDir = filepath.Join("merchant_data")
	//商户资料缓存
	cacheFile = "merchant.db"
	//配置文件名
	configFileName = "merchant.ini"
	//默认配置内容
	defaultConfig = `
# merchant node ID
merchant_node_id = ""
# local node publicKey 
publickey = ""
# local node privateKey 
privatekey = ""
# merchant node api url
merchant_node_url = ""
`
)

//printConfig Print config information
func printConfig() error {

	initConfig()
	//读取配置
	absFile := filepath.Join(merchantDir, configFileName)
	fmt.Printf("-----------------------------------------------------------\n")
	file.PrintFile(absFile)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil

}

//initConfig 初始化配置文件
func initConfig() {

	//读取配置
	absFile := filepath.Join(merchantDir, configFileName)
	if !file.Exists(absFile) {
		file.MkdirAll(merchantDir)
		file.WriteFile(absFile, []byte(defaultConfig), false)
	}

}

//loadConfig 读取配置
func loadConfig() (NodeConfig, error) {

	var (
		c       config.Configer
		err     error
		configs = NodeConfig{}
	)

	//读取配置
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return configs, errors.New("Config is not setup. Please run 'wmd merchant config -i ' ")
	}

	configs.MerchantNodeID = c.String("merchant_node_id")
	configs.LocalPrivateKey = c.String("publickey")
	configs.MerchantNodeURL = c.String("merchant_node_url")
	configs.LocalPrivateKey = c.String("privatekey")
	configs.CacheFile = filepath.Join(merchantDir, cacheFile)

	return configs, nil
}
