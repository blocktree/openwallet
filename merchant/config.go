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

package merchant

import (
	"path/filepath"
	"encoding/json"
	"github.com/blocktree/OpenWallet/common/file"
	"fmt"
	"github.com/astaxie/beego/config"
	"errors"
)

var (

	//商户资料文件夹
	merchantDir = filepath.Join("merchant_data")
	//商户资料缓存
	cacheFile = "merchant.db"
	//配置文件名
	configFileName =  "merchant.ini"
	//默认配置内容
	defaultConfig = `
# merchant node publicKey
nodeKey = ""
# local node publicKey 
publicKey = ""
# local node privateKey 
privateKey = ""
# merchant node api url
merchantNodeURL = ""
# local node id
nodeID = 1
`
)

//newConfigFile 创建配置文件
func newConfigFile(
	nodeKey, publicKey, privateKey string,
	merchantNodeURL string, nodeID int64) (config.Configer, string, error) {

	//	生成配置
	configMap := map[string]interface{}{
		"nodeKey":        nodeKey,
		"publicKey":    publicKey,
		"privateKey":    privateKey,
		"merchantNodeURL":     merchantNodeURL,
		"nodeID":     nodeID,
	}

	filepath.Join()

	bytes, err := json.Marshal(configMap)
	if err != nil {
		return nil, "", err
	}

	//实例化配置
	c, err := config.NewConfigData("json", bytes)
	if err != nil {
		return nil, "", err
	}

	//写入配置到文件
	file.MkdirAll(merchantDir)
	absFile := filepath.Join(merchantDir, configFileName)
	err = c.SaveConfigFile(absFile)
	if err != nil {
		return nil, "", err
	}

	return c, absFile, nil
}

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
		c   config.Configer
		err error
		configs = NodeConfig{}
	)

	//读取配置
	file.MkdirAll(merchantDir)
	absFile := filepath.Join(merchantDir, configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return configs, errors.New("Config is not setup. Please run 'wmd merchant config -i ' ")
	}

	configs.MerchantNodeURL = c.String("merchantNodeURL")
	configs.PublicKey = c.String("publicKey")
	configs.NodeID = c.String("nodeID")
	configs.PrivateKey = c.String("privateKey")
	configs.CacheFile = filepath.Join(merchantDir, cacheFile)

	return configs, nil
}
