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

import "github.com/blocktree/OpenWallet/owtp"

func init() {
	owtp.Debug = true
}

func GetMerchantKeychain() error {

	return nil
}

func InitMerchantKeychainFlow() error {
	return nil
}

//JoinMerchantNodeFlow 连接商户服务节点
func JoinMerchantNodeFlow() error {

	var (
		err error
		c   NodeConfig
	)

	c, err = loadConfig()
	if err != nil {
		return err
	}

	merchantNode, err = NewMerchantNode(c)
	if err != nil {
		return err
	}

	merchantNode.Run()

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

