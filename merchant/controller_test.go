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
	"testing"
	"time"
)

func generateCTX(method string, inputs interface{}) *owtp.Context {
	nonce := uint64(time.Now().Unix())
	ctx := owtp.NewContext(owtp.WSRequest, nonce, method, inputs)
	return ctx
}

func TestSubscribe(t *testing.T) {

	inputs := []Subscription {
		Subscription{Type: 1, Coin:"btc",WalletID:"21212",Version:222},
		Subscription{Type: 1, Coin:"btm",WalletID:"21212",Version:222},
		Subscription{Type: 1, Coin:"ltc",WalletID:"21212",Version:222},
	}

	ctx := generateCTX("subscribe", inputs)

	subscribe(ctx)

	t.Logf("reponse: %v\n",ctx.Resp)
}
