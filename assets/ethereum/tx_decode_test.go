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

package ethereum

import (
	"github.com/tidwall/gjson"
	"testing"
)

func TestNewEthTxExtPara(t *testing.T) {
	extParam := "{\"data\":\"\",\"gasLimit\":\"0.000000000000021\"}"
	extPara := NewEthTxExtPara(gjson.Parse(extParam))
	t.Logf("extPara.GasLimit: %s\n", extPara.GasLimit)
	t.Logf("extPara.Data: %s\n", extPara.Data)
}
