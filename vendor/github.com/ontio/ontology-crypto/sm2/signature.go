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

// Implementation of SM2 signature algorithm

package sm2

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha512"
	"errors"
	"hash"
	"io"
	"math/big"
)

type zr struct {
	io.Reader
}

const (
	aesIV = "IV for <SM2> CTR"

	// DEFAULT_ID is the default user id used in Sign and Verify
	DEFAULT_ID = "1234567812345678"
)

var zeroReader = &zr{}
var one = new(big.Int).SetInt64(1)

type combinedMult interface {
	CombinedMult(bigX, bigY *big.Int, baseScalar, scalar []byte) (x, y *big.Int)
}

func (z *zr) Read(dst []byte) (n int, err error) {
	for i := range dst {
		dst[i] = 0
	}
	return len(dst), nil
}

func randFieldElement(c elliptic.Curve, rand io.Reader) (*big.Int, error) {
	params := c.Params()
	b := make([]byte, params.BitSize/8+8)
	_, err := io.ReadFull(rand, b)
	if err != nil {
		return nil, err
	}

	k := new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(params.N, one)
	n = n.Sub(n, one) //n-2

	// 1 <= k <= n-2
	k.Mod(k, n)
	k.Add(k, one)
	return k, nil
}

// Combine the raw data with user ID, curve parameters and public key
// to generate the signed data used in Sign and Verify
func getZ(msg []byte, pub *ecdsa.PublicKey, userID string, hasher hash.Hash) ([]byte, error) {
	if pub == nil {
		return nil, errors.New("public key should not be nil")
	}

	var c SM2Curve
	if t, ok := pub.Curve.(SM2Curve); !ok {
		return nil, errors.New("the curve type is not SM2Curve")
	} else {
		c = t
	}

	if len(userID) == 0 {
		userID = DEFAULT_ID
	}
	id := []byte(userID)
	len := len(id) * 8
	blen := []byte{byte((len >> 8) & 0xff), byte(len & 0xff)}

	hasher.Reset()
	hasher.Write(blen)
	hasher.Write(id)
	hasher.Write(c.ABytes())
	hasher.Write(c.Params().B.Bytes())
	hasher.Write(c.Params().Gx.Bytes())
	hasher.Write(c.Params().Gy.Bytes())
	hasher.Write(pub.X.Bytes())
	hasher.Write(pub.Y.Bytes())
	h := hasher.Sum(nil)
	return append(h, msg...), nil
}

// Sign generates signature for the input message using the private key and id.
// It returns (r, s) as the signature or error.
func Sign(rand io.Reader, priv *ecdsa.PrivateKey, id string, msg []byte, hasher hash.Hash) (r, s *big.Int, err error) {
	mz, err := getZ(msg, &priv.PublicKey, id, hasher)
	if err != nil {
		return
	}
	hasher.Reset()
	hasher.Write(mz)
	digest := hasher.Sum(nil)

	entropyLen := (priv.Params().BitSize + 7) >> 4
	if entropyLen > 32 {
		entropyLen = 32
	}

	entropy := make([]byte, entropyLen)
	_, err = io.ReadFull(rand, entropy)
	if err != nil {
		return
	}

	priKey := priv.D.Bytes()

	md := sha512.New()
	md.Write(priKey)
	md.Write(entropy)
	md.Write(digest[:])
	key := md.Sum(nil)[:32]

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	cspRng := cipher.StreamReader{
		R: zeroReader,
		S: cipher.NewCTR(block, []byte(aesIV)),
	}

	N := priv.Params().N
	if N.Sign() == 0 {
		err = errors.New("zero parameter")
		return
	}
	var k *big.Int
	e := new(big.Int).SetBytes(digest[:])
	for {
		for {
			k, err = randFieldElement(priv.Curve, cspRng)
			if err != nil {
				r = nil
				err = errors.New("randFieldElement error")
				return
			}

			r, _ = priv.ScalarBaseMult(k.Bytes())
			r.Add(r, e)
			r.Mod(r, N)
			if r.Sign() != 0 {
				break
			}
			if t := new(big.Int).Add(r, k); t.Cmp(N) == 0 {
				break
			}
		}
		D := new(big.Int).SetBytes(priKey)
		rD := new(big.Int).Mul(D, r)
		s = new(big.Int).Sub(k, rD)
		d1 := new(big.Int).Add(D, one)
		d1Inv := new(big.Int).ModInverse(d1, N)
		s.Mul(s, d1Inv)
		s.Mod(s, N)
		if s.Sign() != 0 {
			break
		}
	}

	return
}

// Verify checks whether the input (r, s) is a valid signature for the message.
func Verify(pub *ecdsa.PublicKey, id string, msg []byte, hasher hash.Hash, r, s *big.Int) bool {
	N := pub.Params().N
	if N.Sign() == 0 {
		return false
	}

	t := new(big.Int).Add(r, s)
	t.Mod(t, N)

	var x *big.Int
	if opt, ok := pub.Curve.(combinedMult); ok {
		x, _ = opt.CombinedMult(pub.X, pub.Y, s.Bytes(), t.Bytes())
	} else {
		x1, y1 := pub.ScalarBaseMult(s.Bytes())
		x2, y2 := pub.ScalarMult(pub.X, pub.Y, t.Bytes())
		x, _ = pub.Add(x1, y1, x2, y2)
	}

	mz, err := getZ(msg, pub, id, hasher)
	if err != nil {
		return false
	}

	hasher.Reset()
	hasher.Write(mz)
	digest := hasher.Sum(nil)
	e := new(big.Int).SetBytes(digest[:])
	x.Add(x, e)
	x.Mod(x, N)
	return x.Cmp(r) == 0
}
