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

// Implementation of SM2 public key encryption algorithm

package sm2

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ontio/ontology-crypto/sm3"
)

const DIGESTLEN = 32

func sm3kdf(zInput []byte, kLen int) ([]byte, int) {
	rLen := kLen
	zLen := len(zInput)

	zz := make([]byte, zLen+4)
	for i := 0; i < zLen; i++ {
		zz[i] = zInput[i]
	}

	var pp []byte
	i := 1
	for rLen > 0 {
		zz[zLen] = byte(i>>24) & 0xff
		zz[zLen+1] = byte(i>>16) & 0xFF
		zz[zLen+2] = byte(i>>8) & 0xFF
		zz[zLen+3] = byte(i) & 0xff
		digest := sm3.Sum(zz)

		if rLen >= DIGESTLEN {
			pp = append(pp, digest[:DIGESTLEN]...)
		} else {
			pp = append(pp, digest[:rLen]...)
		}

		rLen -= DIGESTLEN
		i++
	}

	rLen = 0
	pLen := len(pp)
	for i = 0; i < zLen && i < pLen; i++ {
		fmt.Println(" len(pp): ", len(pp), " zLen: ", zLen, " i: ", i)
		if (pp[i] & 0xFF) != 0 {
			rLen = 1
			break
		}
	}

	if rLen > 0 {
		return pp, len(pp)
	} else {
		return nil, 0
	}
}

func Encrypt(pub *ecdsa.PublicKey, data []byte) ([]byte, error) {
	var c SM2Curve
	if t, ok := pub.Curve.(SM2Curve); !ok {
		return nil, errors.New("the curve type is not SM2Curve")
	} else {
		c = t
	}

	encryptData := make([]byte, 96+len(data))

	hash := sm3.Sum(data)
	entropyLen := (c.Params().BitSize + 7) / 16
	if entropyLen > 32 {
		entropyLen = 32
	}
	entropy := make([]byte, entropyLen)
	_, err := io.ReadFull(rand.Reader, entropy)
	if err != nil {
		return nil, err
	}

	md := sha512.New()
	md.Write(pub.X.Bytes())
	md.Write(entropy)
	md.Write(hash[:])
	key := md.Sum(nil)[:32]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cspRng := cipher.StreamReader{
		R: zeroReader,
		S: cipher.NewCTR(block, []byte(aesIV)),
	}

	N := c.Params().N
	if N.Sign() == 0 {
		return nil, errors.New("zero parameter")
	}

	var k, x2, y2 *big.Int
	c2 := make([]byte, len(data))

	for {
		k, err = randFieldElement(c, cspRng)
		if err != nil {
			return nil, errors.New("randFieldElement error")
		}

		x1, y1 := c.ScalarBaseMult(k.Bytes())
		copy(encryptData[32-len(x1.Bytes()):], x1.Bytes())
		copy(encryptData[64-len(y1.Bytes()):], y1.Bytes())

		x2, y2 = c.ScalarMult(pub.X, pub.Y, k.Bytes())
		x2y2 := make([]byte, 64)
		copy(x2y2[32-len(x2.Bytes()):], x2.Bytes())
		copy(x2y2[64-len(y2.Bytes()):], y2.Bytes())

		t, ret := sm3kdf(x2y2, len(data))
		if ret == 0 {
			continue
		}

		for i := 0; i < len(data); i++ {
			c2[i] = data[i] ^ t[i]
		}
		break
	}

	tmp := make([]byte, 32+len(data)+32)
	copy(tmp[32-len(x2.Bytes()):], x2.Bytes())
	copy(tmp[32:], data)
	copy(tmp[len(tmp)-32:], y2.Bytes())
	c3 := sm3.Sum(tmp)

	copy(encryptData[64:], c3[:])
	copy(encryptData[96:], c2)

	return encryptData, nil
}

func Decrypt(priv *ecdsa.PrivateKey, encryptData []byte) ([]byte, error) {
	var c SM2Curve
	if t, ok := priv.Curve.(SM2Curve); !ok {
		return nil, errors.New("the curve type is not SM2Curve")
	} else {
		c = t
	}

	x1 := new(big.Int).SetBytes(encryptData[:32])
	y1 := new(big.Int).SetBytes(encryptData[32:64])

	x2, y2 := c.ScalarMult(x1, y1, priv.D.Bytes())
	c2 := make([]byte, 64)

	copy(c2[32-len(x2.Bytes()):], x2.Bytes())
	copy(c2[64-len(y2.Bytes()):], y2.Bytes())

	MsgLen := len(encryptData) - 96
	t, ret := sm3kdf(c2, MsgLen)
	if ret == 0 {
		return nil, errors.New("KDF calc error")
	}

	Msg := make([]byte, MsgLen)
	for i := 0; i < MsgLen; i++ {
		Msg[i] = encryptData[96:][i] ^ t[i]
	}

	tmp := make([]byte, 32+len(Msg)+32)
	copy(tmp[32-len(x2.Bytes()):], x2.Bytes())
	copy(tmp[32:], Msg)
	copy(tmp[len(tmp)-32:], y2.Bytes())
	c3 := sm3.Sum(tmp)

	if 0 != bytes.Compare(c3[:], encryptData[64:96]) {
		return Msg, errors.New("Hash not match!")
	}
	return Msg, nil
}
