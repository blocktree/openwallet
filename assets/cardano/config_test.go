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

package cardano

import "testing"

func TestIsExistConfig(t *testing.T) {
	flag := isExistConfigFile()
	t.Logf("is exist: %v", flag)
}

func TestNewConfigFile(t *testing.T) {

	var (
		apiURL     = "https://192.168.2.224:10026/api/"
		walletPath = "/home/ada/cardano-sl/state-wallet-mainnet"
		//汇总阀值
		threshold uint64 = 9 * 1000000
		//最小转账额度
		minSendAmount uint64 = 1 * 1000000
		//最小矿工费
		minFees uint64 = 0.3 * 1000000
		//汇总地址
		sumAddress = "DdzFFzCqrhsgCwX9VGWZAdAnPfeVwUzoPDAqQJfXE3DxKKFCTYmtGD9CrKjvGu7VjebGoCqPHN7DtkF1VhEvJbPhg2BrfhT5hkyBvjvZ"
	)

	_, _, err := newConfigFile(apiURL, walletPath, sumAddress, threshold, minSendAmount, minFees)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

}

func TestInitConfigFlow(t *testing.T) {
	InitConfigFlow()
}