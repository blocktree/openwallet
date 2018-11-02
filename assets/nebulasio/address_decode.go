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

package nebulasio

import (
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"github.com/blocktree/go-owcrypt"
)

func init() {

}

//var (
//	AddressDecoder = &openwallet.AddressDecoder{
//		PrivateKeyToWIF:    PrivateKeyToWIF,
//		PublicKeyToAddress: PublicKeyToAddress,
//		WIFToPrivateKey:    WIFToPrivateKey,
//		RedeemScriptToAddress: RedeemScriptToAddress,
//	}
//)

//地址：n1Fvg4DvRhBE4YuC2d884TKQnBuksJYXem4
//私钥：0x57,0xa5,0x27,0xaa,0x5e,0xc8,0xdd,0xa9,0x6e,0x4c,0xbe,0x16,0x29,0xcf,0xf1,0x08,0x6d,0x80,0xd4,0xd3,0xaf,0x67,0x14,0x58,0x97,0x45,0x39,0x64,0x16,0xce,0x7f,0xfd
//公钥：0x02,0x43,0xe7,0x98,0x6f,0x5a,0xb0,0xcd,0xdb,0x95,0xb2,0xf9,0x4a,0x12,0x5e,0x8a,0x67,0x0e,0x49,0x94,0x0f,0xa7,0xf9,0xff,0x65,0x7c,0xb0,0xbb,0xaf,0x15,0xcb,0x3c,0x57
type addressDecoder struct {
	wm *WalletManager //钱包管理者
}

//NewAddressDecoder 地址解析器
func NewAddressDecoder(wm *WalletManager) *addressDecoder {
	decoder := addressDecoder{}
	decoder.wm = wm
	return &decoder
}

//PrivateKeyToWIF 私钥转WIF
func (decoder *addressDecoder) PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error) {
/*
	cfg := addressEncoder.NAS_AccountAddress

	//privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), priv)
	//wif, err := btcutil.NewWIF(privateKey, &cfg, true)
	//if err != nil {
	//	return "", err
	//}

	wif := addressEncoder.AddressEncode(priv, cfg)

	return wif, nil*/
	return "", nil
}

//PublicKeyToAddress 公钥转地址,入参pub为压缩33位
func (decoder *addressDecoder) PublicKeyToAddress(pub []byte, isTestnet bool) (string, error) {

	PublicKey_decode := pub
	// 如入参是33字节的压缩公钥进行解压成65字节
	// 04730923f3f99eb587cbbcfa4876a9be518d8893a2106ebb93e40def9af95c308ce09de098076869d067c3e82564e673de3f965585d2466383349b20bd8bface0a
	if(len(pub) == 33) {
		PublicKey_decode = owcrypt.PointDecompress(pub, owcrypt.ECC_CURVE_SECP256K1)
		//fmt.Printf("PublicKey_decode=%x\n", PublicKey_decode)
	}

	//对于有些币种只对整个解压的公钥[1:65]或者压缩的33字节公钥算hash之后编码得到地址
	//NAS需要对整个解压的公钥算hash之后编码得到地址 n1VC5UsE2mVRYMuctV66BtqFkdi1KCUZodY,
	//因为官方节点验签规则：根据广播交易单对签名结果恢复得到65字节公钥，再去算签名的地址,与from地址进行对比，一致则验签通过
	cfg := addressEncoder.NAS_AccountAddress
	pkHash := owcrypt.Hash(PublicKey_decode, 20, owcrypt.HASH_ALG_SHA3_256_RIPEMD160)
	address := addressEncoder.AddressEncode(pkHash, cfg)

	return address, nil
}

//WIFToPrivateKey WIF转私钥
func (decoder *addressDecoder) WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error) {

/*	cfg := addressEncoder.NAS_AccountAddress

	priv, err := addressEncoder.AddressDecode(wif, cfg)
	if err != nil {
		return nil, err
	}

	return priv, err*/

	return nil,nil
}

//RedeemScriptToAddress 多重签名赎回脚本转地址
func (decoder *addressDecoder) RedeemScriptToAddress(pubs [][]byte, required uint64, isTestnet bool) (string, error) {

	cfg := addressEncoder.NAS_AccountAddress

	redeemScript := make([]byte, 0)

	for _, pub := range pubs {
		redeemScript = append(redeemScript, pub...)
	}

	//wjq
	pkHash := owcrypt.Hash(redeemScript, 20, owcrypt.HASH_ALG_SHA3_256_RIPEMD160)
	address := addressEncoder.AddressEncode(pkHash, cfg)

	return address, nil
}

