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

package signature

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"io"
	"math/big"

	"golang.org/x/crypto/ed25519"

	"github.com/ontio/ontology-crypto/ec"
	"github.com/ontio/ontology-crypto/sm2"
)

type Signature struct {
	Scheme SignatureScheme
	Value  interface{}
}

type DSASignature struct {
	R, S *big.Int
}

type SM2Signature struct {
	DSASignature
	ID string
}

// Sign generates the signature for the input message @msg, using private key
// @pri and the signature scheme @scheme.
//
// Some signature scheme may use extra parameters, which could be inputted via
// the last argument @opt:
// - SM2 signature needs the user ID (string). If it is an empty string, the
//   default ID ("1234567812345678") would be used.
func Sign(scheme SignatureScheme, pri crypto.PrivateKey, msg []byte, opt interface{}) (sig *Signature, err error) {
	var res Signature
	res.Scheme = scheme
	switch key := pri.(type) {
	case *ec.PrivateKey:
		hasher := GetHash(scheme)
		if hasher == nil {
			err = errors.New("signing failed: unknown scheme")
			return
		}

		if scheme == SM3withSM2 {
			id := ""
			if opt, ok := opt.(string); ok {
				id = opt
			}
			r, s, err0 := sm2.Sign(rand.Reader, key.PrivateKey, id, msg, hasher)
			if err0 != nil {
				err = err0
				return
			}
			res.Value = &SM2Signature{
				ID:           id,
				DSASignature: DSASignature{R: r, S: s},
			}
		} else if scheme == SHA224withECDSA ||
			scheme == SHA256withECDSA ||
			scheme == SHA384withECDSA ||
			scheme == SHA512withECDSA ||
			scheme == SHA3_224withECDSA ||
			scheme == SHA3_256withECDSA ||
			scheme == SHA3_384withECDSA ||
			scheme == SHA3_512withECDSA ||
			scheme == RIPEMD160withECDSA {

			hasher.Write(msg)
			digest := hasher.Sum(nil)
			r, s, err0 := ecdsa.Sign(rand.Reader, key.PrivateKey, digest)
			if err0 != nil {
				err = err0
				return
			}
			res.Value = &DSASignature{R: r, S: s}
		} else {
			err = errors.New("signing failed: unmatched signature scheme and private key")
			return
		}

	case ed25519.PrivateKey:
		if scheme != SHA512withEDDSA {
			err = errors.New("signing failed: unmatched signature scheme and private key")
			return
		}
		res.Value = ed25519.Sign(key, msg)

	default:
		err = errors.New("signing failed: unknown type of private key")
		return
	}

	sig = &res
	return
}

// Verify checks whether @sig is a valid signature for @msg with the public key
// @pub, and return true/false as the result.
func Verify(pub crypto.PublicKey, msg []byte, sig *Signature) bool {
	if len(msg) == 0 || sig == nil {
		return false
	}

	h := GetHash(sig.Scheme)
	if h == nil {
		return false
	}

	res := false

	switch key := pub.(type) {
	case *ec.PublicKey:
		switch sig.Scheme {
		case SHA224withECDSA, SHA256withECDSA, SHA384withECDSA, SHA512withECDSA, SHA3_224withECDSA, SHA3_256withECDSA, SHA3_384withECDSA, SHA3_512withECDSA, RIPEMD160withECDSA:
			if v, ok := sig.Value.(*DSASignature); ok {
				h.Write(msg)
				digest := h.Sum(nil)
				res = ecdsa.Verify(key.PublicKey, digest, v.R, v.S)
			}
		case SM3withSM2:
			if v, ok := sig.Value.(*SM2Signature); ok {
				res = sm2.Verify(key.PublicKey, v.ID, msg, h, v.R, v.S)
			}
		}
	case ed25519.PublicKey:
		if sig.Scheme == SHA512withEDDSA {
			v := sig.Value.([]byte)
			res = ed25519.Verify(key, msg, v)
		}
	}

	return res
}

// Serialize the signature object to byte array as the following format:
//
//     |---------------------------|-----------------|
//     | signature_scheme (1 byte) | signature_data  |
//     |---------------------------|-----------------|
//
// signature_data differs in the signature algorithm.
// - For ECDSA, it is the concatenation of two byte arrays of equal length
//   converted from R and S.
// - For SM2, it starts with the user ID (empty if not specified) with a `0`
//   as termination, and followed by the R and S data as described in ECDSA.
// - For EdDSA, it is just the signature data which could be handled by package
//   ed25519.
func Serialize(sig *Signature) ([]byte, error) {
	if sig == nil {
		return nil, errors.New("failed serializing signature: input is nil")
	}

	var buf bytes.Buffer
	buf.WriteByte(byte(sig.Scheme))
	switch v := sig.Value.(type) {
	case *DSASignature:
		if sig.Scheme != SHA224withECDSA &&
			sig.Scheme != SHA256withECDSA &&
			sig.Scheme != SHA384withECDSA &&
			sig.Scheme != SHA512withECDSA &&
			sig.Scheme != SHA3_224withECDSA &&
			sig.Scheme != SHA3_256withECDSA &&
			sig.Scheme != SHA3_384withECDSA &&
			sig.Scheme != SHA3_512withECDSA &&
			sig.Scheme != RIPEMD160withECDSA {
			return nil, errors.New("failed serializing signature: unmatched signature scheme and value")
		}

		serializeDSA(v, &buf)

	case *SM2Signature:
		if sig.Scheme != SM3withSM2 {
			return nil, errors.New("failed serializing signature: unmatched signature scheme and value")
		}
		buf.Write([]byte(v.ID))
		buf.WriteByte(byte(0))
		serializeDSA(&v.DSASignature, &buf)
	case []byte:
		buf.Write(v)
	default:
		return nil, errors.New("failed serializing signature: unrecognized signature type")
	}

	res := buf.Bytes()

	// Treat SHA256withECDSA as a special case, using the signature
	// data directly without the signature scheme.
	if sig.Scheme == SHA256withECDSA && len(res) == 65 {
		res = res[1:]
	}

	return res, nil
}

// Deserialize the input data into a Signature object.
func Deserialize(buf []byte) (*Signature, error) {
	e := "failed deserializing signature: "
	if buf == nil || len(buf) < 2 {
		return nil, errors.New(e + "invalid argument")
	}

	var sig Signature
	var data []byte
	if len(buf) == 64 {
		data = buf
		sig.Scheme = SHA256withECDSA
	} else {
		data = buf[1:]
		sig.Scheme = SignatureScheme(buf[0])
	}

	switch sig.Scheme {
	case SHA224withECDSA, SHA256withECDSA, SHA384withECDSA, SHA512withECDSA,
		SHA3_224withECDSA, SHA3_256withECDSA, SHA3_384withECDSA, SHA3_512withECDSA,
		RIPEMD160withECDSA:
		dsa, err := deserializeDSA(data)
		if err != nil {
			return nil, errors.New(e + err.Error())
		}
		sig.Value = dsa
	case SM3withSM2:
		i := 0
		for i < len(data) && data[i] != 0 {
			i++
		}
		if i >= len(data) {
			return nil, errors.New(e + "invalid format")
		}
		id := string(data[0:i])
		dsa, err := deserializeDSA(data[i+1:])
		if err != nil {
			return nil, errors.New(e + err.Error())
		}
		sig.Value = &SM2Signature{ID: id, DSASignature: *dsa}
	case SHA512withEDDSA:
		sig.Value = data
	default:
		return nil, errors.New(e + "unknown signature scheme")
	}
	return &sig, nil
}

func serializeDSA(sig *DSASignature, w io.Writer) {
	if sig == nil || sig.R == nil || sig.S == nil {
		panic("serializeDSA: invalid argument")
	}

	r := sig.R.Bytes()
	s := sig.S.Bytes()
	lr := len(r)
	ls := len(s)
	if lr < ls {
		w.Write(make([]byte, ls-lr))
	}
	w.Write(r)
	if ls < lr {
		w.Write(make([]byte, lr-ls))
	}
	w.Write(s)
}

func deserializeDSA(buf []byte) (*DSASignature, error) {
	if buf == nil {
		panic("deserializeDSA: invalid argument")
	}

	length := len(buf)
	if length&1 != 0 {
		return nil, errors.New("invalid length")
	}

	return &DSASignature{
		R: new(big.Int).SetBytes(buf[0 : length/2]),
		S: new(big.Int).SetBytes(buf[length/2:]),
	}, nil
}
