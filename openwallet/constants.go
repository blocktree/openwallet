/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwallet

//钱包类型
type WalletType = uint

const (
	WalletTypeSingle = 0
	WalletTypeMulti = 1
)

const (

	/// 私钥字节长度
	PrivateKeyLength = 32

	/// 公钥字节长度（压缩）
	PublicKeyLengthCompressed = 33

	/// 公钥字节长度（未压缩）
	PublicKeyLengthUncompressed = 65

)


