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

	TOADDRESS    string = "TWVRXXN5tsggjUCDmqbJ4KxPdJKQiynaG6"                               // account: t2
	OWNERADDRESS string = "TQ1TiUzStbSLdEtACUDmzfMDpWUyo8cyCf"                               // account: t1
	PRIVATEKEY   string = "9e9fa25c9d70fecc91c90d23b55daffa2f5f23ffa9eeca823260e50e544cf7be" // account: t1
	AMOUNT       int64  = 1

	TXRAW    string = "0a7e0a02d5842208a43b6160eaa543f840f8ffb4a9e12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d"
	TXSIGNED string = "0a7e0a02d5842208a43b6160eaa543f840f8ffb4a9e12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d1241915254f565f20a327047e59327bf7f0b9e5600c09a321e6da45de4be4158f1f14bbf520ac1a0d71da10696ee80dfa5860086a6e470c96e7a275c4433616b8beb00"
)

func init() {

	tw = NewWalletManager()

	// tw.Config.ServerAPI = "http://127.0.0.1:28890"
	// tw.Config.ServerAPI = "http://127.0.0.1:18890"
	tw.Config.ServerAPI = "http://192.168.2.194:18890"
	// tw.Config.ServerAPI = "http://192.168.2.194:28890"
	// tw.Config.RpcUser = "walletUser"
	// tw.Config.RpcPassword = "walletPassword2017"
	// token := BasicAuth(tw.Config.RpcUser, tw.Config.RpcPassword)
	tw.WalletClient = NewClient(tw.Config.ServerAPI, "", true)
}
