/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

//Ontolog sdk in golang. Using for operation with ontology
package ontology_go_sdk

import (
	"fmt"
	"github.com/ontio/ontology-go-sdk/rest"
	"github.com/ontio/ontology-go-sdk/rpc"
	"github.com/ontio/ontology-go-sdk/utils"
	"github.com/ontio/ontology/account"
)

//OntologySdk is the main struct for user
type OntologySdk struct {
	//Rpc client used the rpc api of ontology
	Rpc *rpc.RpcClient
	//Rest client used the rest api of ontology
	Rest *rest.RestClient
}

//NewOntologySdk return OntologySdk.
func NewOntologySdk() *OntologySdk {
	return &OntologySdk{
		Rpc:  rpc.NewRpcClient(),
		Rest: rest.NewRestClient(),
	}
}

//OpenOrCreateWallet return a wllet instance.If the wallet is exist, just open it. if not, then create and open.
func (this *OntologySdk) OpenOrCreateWallet(walletFile string) (account.Client, error) {
	if utils.IsFileExist(walletFile) {
		return this.OpenWallet(walletFile)
	} else {
		return this.CreateWallet(walletFile)
	}
}

//CreateWallet return a new wallet
func (this *OntologySdk) CreateWallet(walletFile string) (account.Client, error) {
	if utils.IsFileExist(walletFile) {
		return nil, fmt.Errorf("wallet:%s has already exist", walletFile)
	}
	return account.Open(walletFile)
}

//OpenWallet return a wallet instance
func (this *OntologySdk) OpenWallet(walletFile string) (account.Client, error) {
	return account.Open(walletFile)
}
