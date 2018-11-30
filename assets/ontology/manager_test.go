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

package ontology

var (
	tw *WalletManager
)

func init() {

	tw = NewWalletManager()

	tw.Config.ServerAPI = "http://127.0.0.1:20336"
	tw.Config.RpcUser = ""
	tw.Config.RpcPassword = ""
	token := BasicAuth(tw.Config.RpcUser, tw.Config.RpcPassword)
	tw.WalletClient = NewClient(tw.Config.ServerAPI, token, false)

	//explorerURL := "http://192.168.32.107:20003/insight-api/"
	//tw.ExplorerClient = NewExplorer(explorerURL, true)

	localURL := "http://localhost:20334/api/v1/"
	tw.LocalClient = NewLocal(localURL, true)

	tw.Config.RPCServerType = RPCServerCore
}
