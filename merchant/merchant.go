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
	"github.com/blocktree/OpenWallet/owtp"
	"github.com/pkg/errors"
	"log"
	"path/filepath"
)


func GetMerchantKeychain() error {

	return nil
}

func InitMerchantKeychainFlow() error {
	return nil
}

//JoinMerchantNodeFlow 连接商户服务节点
func JoinMerchantNodeFlow() error {

	var (
		err        error
		auth       *owtp.OWTPAuth
	)

	err = loadConfig()
	if err != nil {
		return err
	}

	if len(merchantNodeURL) == 0 {
		return errors.New("merchant node url is not configed!")
	}

	if merchantNode != nil {
		merchantNode.Close()
	}

	//授权配置
	auth, err = owtp.NewOWTPAuth(
		nodeKey,
		publicKey,
		privateKey,
		true,
		filepath.Join(merchantDir, cacheFile))

	if err != nil {
		return err
	}

	//创建节点，连接商户
	merchantNode = owtp.NewOWTPNode(nodeID, merchantNodeURL, auth)
	//设置路由
	setupRouter(merchantNode)

	//断开连接后，重新连接
	merchantNode.SetCloseHandler(func(n *owtp.OWTPNode) {
		log.Printf("merchantNode disconnect. \n")
		disconnected <- struct{}{}
	})

	//商户连接运行
	run()

	return nil
}

func ConfigMerchantFlow() error {
	initConfig()
	return nil
}

func ShowMechantConfig() error {
	printConfig()
	return nil
}

