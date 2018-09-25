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

package tron

var (
	tw *WalletManager

	to_address    string = "TWzVsJyUvTjVRwDTRNxUsLCjH9KY9gQgk2"
	owner_address string = "TSdXzXKSQ3RQzQ5Ge8TiYfMQEjofSVQ8ax"
	privateKey    string = "57ddf47e1b1aaabf244d1429f300a001e16fe407d9d4c9b6e43d19b128b4b442"
	amount        uint64 = 1

	txID string = "d5ec749ecc2a615399d8a6c864ea4c74ff9f523c2be0e341ac9be5d47d7c2d62"
)

func init() {

	tw = NewWalletManager()

	// tw.Config.ServerAPI = "http://127.0.0.1:28890"
	tw.Config.ServerAPI = "http://127.0.0.1:18890"
	// tw.Config.RpcUser = "walletUser"
	// tw.Config.RpcPassword = "walletPassword2017"
	// token := BasicAuth(tw.Config.RpcUser, tw.Config.RpcPassword)
	tw.WalletClient = NewClient(tw.Config.ServerAPI, "", true)
}
