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

import (
	"github.com/tidwall/gjson"

	"github.com/tronprotocol/grpc-gateway/api"
)

// wallet/listwitnesses Function：Query the list of Super Representatives demo: curl -X POSThttp://127.0.0.1:8090/wallet/listwitnesses Parameters：None Return value：List of all Super Representatives
func (wm *WalletManager) ListWitnesses() (witnesses *api.WitnessList, err error) {

	r, err := wm.WalletClient.Call2("/wallet/listwitnesses", nil)
	if err != nil {
		return nil, err
	}

	witnesses = &api.WitnessList{}
	if err := gjson.Unmarshal(r, witnesses); err != nil {
		return nil, err
	}

	return witnesses, nil
}

// wallet/listnodes
// Function：List the nodes which the api fullnode is connecting on the network
// demo: curl -X POST http://127.0.0.1:8090/wallet/listnodes
// Parameters：None
// Return value：List of nodes
func (wm *WalletManager) ListNodes() (nodes *api.NodeList, err error) {

	// request := []interface{}{
	// 	// walletID,
	// }

	r, err := wm.WalletClient.Call2("/wallet/listnodes", nil)
	if err != nil {
		return nil, err
	}

	nodes = &api.NodeList{}
	if err := gjson.Unmarshal(r, nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}
