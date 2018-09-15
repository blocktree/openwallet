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
	"fmt"
	"math/big"

	. "github.com/blocktree/go-OWCrypt"
)

//--------------------------- ETH signature start--------------------------------
/*
@brife: according to RFC6979 standard
@function: init HMAC
@paramter[in]key pointer to private key ||hash(message) in signature procedure
@paramter[in]keylen is the byte length of key
@paramter[out]the first return value pointer to k||v(type is []byte)
@paramter[out]the second return value is retry(type is int)
*/
func HMAC_RFC6979_init(key []byte, keylen int) ([]byte, int) {
	k := make([]byte, 32)
	v := make([]byte, 32)
	out := make([]byte, 64)
	tempbuf := make([]byte, 33+keylen)
	//step b in RFC6979
	//copy(v,0x01,32)
	for i := 0; i < 32; i++ {
		v[i] = 0x1
	}
	//step c in RFC6979
	for i := 0; i < 32; i++ {
		k[i] = 0x0
	}
	//step d in RFC6979
	//copy(tempbuf,v,32)
	copy(tempbuf[:32], v[:])
	//memcpy(tempbuf+32,0x0,1)
	tempbuf[32] = 0
	//memcpy(tempbuf+33,key,keylen)
	copy(tempbuf[33:33+keylen], key[:])
	//kpointer :=(*C.uchar)(unsafe.Pointer(&k[0]))
	//tempointer :=(*C.uchar)(unsafe.Pointer(&tempbuf[0]))
	//vpointer :=(*C.uchar)(unsafe.Pointer(&v[0]))
	k = Hmac(k, tempbuf, HMAC_SHA256_ALG)

	//step e in RFC6979
	v = Hmac(k, v, HMAC_SHA256_ALG)

	//step f in RFC6979
	//memcpy(tempbuf,v,32)
	copy(tempbuf[:32], v[:])
	//memset(tempbuf+32,0x01,1)
	tempbuf[32] = 0x01
	k = Hmac(k, tempbuf, HMAC_SHA256_ALG)
	//step g in RFC6979
	v = Hmac(k, v, HMAC_SHA256_ALG)
	retry := 0
	//append(k,v)
	copy(out[:32], k[:])
	copy(out[32:64], v[:])
	//返回k||v,retry
	return out, retry
}

/*
@brife: according to RFC6979 standard
@paramter[in]k pointer to k generates in HMAC_RFC6979_init()
@paramter[in]v ipointer to v generates in HMAC_RFC6979_init()
@paramter[in]retry denotes the retry in HMAC_RFC6979_init()
@paramter[out]the first return value pointer to nounce(type is []byte)
@paramter[out]the second return value is retry(type is int)
*/
func HMAC_RFC6979_gnerate(k, v []byte, retry, nouncelen int) ([]byte, int) {
	nounce := make([]byte, nouncelen)
	j := 0
	if retry == 1 {
		tempbuf := make([]byte, 33)
		//memcpy(tempbuf,v,32)
		copy(tempbuf[:32], v[:])
		//memset(tempbuf,0,1)
		tempbuf[32] = 0
		k = Hmac(k, tempbuf, HMAC_SHA256_ALG)
		v = Hmac(k, v, HMAC_SHA256_ALG)
	}
	for i := 0; i < nouncelen; i += 32 {
		v = Hmac(k[:], v[:], HMAC_SHA256_ALG)
		//memcpy(nounce+i,v,32)
		//copy(nounce[i*32:(i+1)*32],v[:])
		copy(nounce[(j*32):((j+1)*32)], v[:])
		j++
	}
	retry = 1
	return nounce, retry
}

func nonce_function_rfc6979(msg, key, algo, extradata []byte, counter int) []byte {
	keydata := make([]byte, 112)
	nounce := make([]byte, (counter+1)*32)
	//memcpy(keydata,key,32)
	copy(keydata[:32], key[:])
	//memcpy(keydata + 32,msg,32)
	copy(keydata[32:64], msg[:])
	keylen := 64
	if extradata != nil {
		//memcpy(keydata+64,msg,32)
		copy(keydata[64:96], extradata[:])
		keylen += 32
	}
	if algo != nil {
		//memcpy(keydata+keylen,algo,16)
		copy(keydata[keylen:keylen+16], algo[:])
	}
	ret, retry := HMAC_RFC6979_init(keydata, keylen)

	for i := 0; i <= counter; i++ {
		ReTry := retry
		temp, retry := HMAC_RFC6979_gnerate(ret[:32], ret[32:], ReTry, 32)
		copy(nounce[i*32:(i+1)*32], temp[:])
		ReTry = retry
	}

	return nounce

}

/*
@function:ETH signature
@paramter[in]prikey pointer to private key
@paramter[in]hash pointer to the hash of message(Transaction txt)
@parameter[out]the first part is signature(r||s||v,total 65 byte);
the second part
*/
func ETHsignatureInner(prikey []byte, hash []byte, counter int) ([]byte, uint16) {
	signature := make([]byte, 65)
	var recid byte
	//var ret uint16
	if len(hash) != 32 {
		fmt.Sprintln("hash is required to be exactly 32 bytes")
		return nil, FAILURE
	}
	nounce := nonce_function_rfc6979(hash, prikey, nil, nil, 0)
	ret := PreprocessRandomNum(nounce)
	if ret != SUCCESS {
		return nil, ret
	}

	//外部传入随机数，外部已经计算哈希值
	sig, ret := Signature(prikey, nil, 0, hash, 32, ECC_CURVE_SECP256K1|NOUNCE_OUTSIDE_FLAG|HASH_OUTSIDE_FLAG)
	curveOrder := new(big.Int).SetBytes(GetCurveOrder(ECC_CURVE_SECP256K1))
	halfcurveorder := big.NewInt(0)
	s := new(big.Int).SetBytes(sig[32:64])
	divider := big.NewInt(2)
	halfcurveorder.Div(curveOrder, divider)
	sign := s.Cmp(halfcurveorder)
	if sign > 0 {
		s.Sub(curveOrder, s)
		copy(sig[32:64], s.Bytes())
	}
	//判断[nounce]G(G is base point) Y-coordinate 的奇偶性,如果为奇数，recid=0x0;如果为奇数，recid=0x01.
	//yPoint, ret1 := GenPubkey(nounce, ECC_CURVE_SECP256K1)[32:]
	yPoint, ret1 := GenPubkey(nounce, ECC_CURVE_SECP256K1)
	if ret1 != SUCCESS {
		return nil, ret1
	}
	if yPoint[63]%2 == 1 {
		recid = 0x01
	} else {
		recid = 0x00
	}
	copy(signature[:64], sig[:])
	signature[64] = recid
	return signature, ret
}

func ETHsignature(prikey []byte, hash []byte) ([]byte, uint16) {
	pri := make([]byte, 65)
	var ret uint16
	counter := 0
	for {
		pri, ret = ETHsignatureInner(prikey, hash, counter)
		if ret == SUCCESS {
			break
		}
		counter++
	}
	return pri, ret
}
