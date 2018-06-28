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

// Package keypair implements asymmetric key pair generation and some related
// functions.
//
// Multiple types of key pair supported:
//     ECDSA
//     SM2
//     EdDSA
//
package keypair

import (
	"bytes"
	"crypto"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	base58 "github.com/itchyny/base58-go"
	"github.com/ontio/ontology-crypto/ec"

	"golang.org/x/crypto/ed25519"
)

type PublicKey crypto.PublicKey

type PrivateKey interface {
	crypto.PrivateKey
	Public() crypto.PublicKey
}

type KeyType byte

// Supported key types
const (
	PK_ECDSA KeyType = 0x12
	PK_SM2   KeyType = 0x13
	PK_EDDSA KeyType = 0x14

	PK_P256_E KeyType = 0x02
	PK_P256_O KeyType = 0x03
)

const err_generate = "key pair generation failed, "

// GenerateKeyPair generates a pair of private and public keys in type t.
// opts is the necessary parameter(s), which is defined by the key type:
//     ECDSA: a byte specifies the elliptic curve, which defined in package ec
//     SM2:   same as ECDSA
//     EdDSA: a byte specifies the curve, only ED25519 supported currently.
func GenerateKeyPair(t KeyType, opts interface{}) (PrivateKey, PublicKey, error) {
	switch t {
	case PK_ECDSA, PK_SM2:
		param, ok := opts.(byte)
		if !ok {
			return nil, nil, errors.New(err_generate + "invalid EC options, 1 byte curve label excepted")
		}
		c, err := GetCurve(param)
		if err != nil {
			return nil, nil, errors.New(err_generate + err.Error())
		}

		if t == PK_ECDSA {
			return ec.GenerateECKeyPair(c, rand.Reader, ec.ECDSA)
		} else {
			return ec.GenerateECKeyPair(c, rand.Reader, ec.SM2)
		}

	case PK_EDDSA:
		param, ok := opts.(byte)
		if !ok {
			return nil, nil, errors.New(err_generate + "invalid EdDSA option")
		}

		if param == ED25519 {
			pub, pri, err := ed25519.GenerateKey(rand.Reader)
			return pri, pub, err
		} else {
			return nil, nil, errors.New(err_generate + "unsupported EdDSA scheme")
		}
	default:
		return nil, nil, errors.New(err_generate + "unknown algorithm")
	}
}

func GetKeyType(p PublicKey) KeyType {
	switch t := p.(type) {
	case *ec.PublicKey:
		switch t.Algorithm {
		case ec.ECDSA:
			return PK_ECDSA
		case ec.SM2:
			return PK_SM2
		default:
			panic("unknown public key type")
		}
	case ed25519.PublicKey:
		return PK_EDDSA
	default:
		panic("unknown public key type")
	}
}

// SerializePublicKey serializes the public key to a byte sequence as the
// following format:
//         |--------------------|-----------------|
//         | algorithm (1 byte) | public_key_data |
//         |--------------------|-----------------|
//
// public_key_data differs in the algorithm:
//
// - ECDSA & SM2
//         |----------------|--------------------|
//         | curve (1 byte) | encoded_public_key |
//         |----------------|--------------------|
//   encoded_public_key is the public key encoded in compression mode.
//
// - EdDSA
//   Since only Ed25519 supported currently, it is just the 1 byte curve label
//   followed by the byte sequence which could be handled as public key in
//   package ed25519.
//
// ECDSA public key with NIST P-256 curve is treated as a special case, which
// just use the encoded data as the serialization and starts with 0x02 or 0x03,
// with no flags ahead.
//
// This function will panic if error occurs.
func SerializePublicKey(key PublicKey) []byte {
	var buf bytes.Buffer
	switch t := key.(type) {
	case *ec.PublicKey:
		switch t.Algorithm {
		case ec.ECDSA:
			// Take P-256 as a special case
			if t.Params().Name == elliptic.P256().Params().Name {
				return ec.EncodePublicKey(t.PublicKey, true)
			}
			buf.WriteByte(byte(PK_ECDSA))
		case ec.SM2:
			buf.WriteByte(byte(PK_SM2))
		}
		label, err := GetCurveLabel(t.Curve)
		if err != nil {
			panic(err)
		}
		buf.WriteByte(label)
		buf.Write(ec.EncodePublicKey(t.PublicKey, true))
	case ed25519.PublicKey:
		buf.WriteByte(byte(PK_EDDSA))
		buf.WriteByte(ED25519)
		buf.Write([]byte(t))
	default:
		panic("unknown public key type")
	}

	return buf.Bytes()
}

// DeserializePublicKey parse the byte sequencce to a public key.
func DeserializePublicKey(data []byte) (PublicKey, error) {
	if len(data) <= 3 {
		return nil, errors.New("too short pubkey")
	}
	switch KeyType(data[0]) {
	case PK_ECDSA, PK_SM2:
		c, err := GetCurve(data[1])
		if err != nil {
			return nil, err
		}
		pub, err := ec.DecodePublicKey(data[2:], c)
		if err != nil {
			return nil, err
		}
		pk := &ec.PublicKey{PublicKey: pub}
		switch KeyType(data[0]) {
		case PK_ECDSA:
			pk.Algorithm = ec.ECDSA
		case PK_SM2:
			pk.Algorithm = ec.SM2
		default:
			return nil, errors.New("deserializing public key failed: unknown EC algorithm")
		}

		return pk, nil

	case PK_EDDSA:
		if data[1] == ED25519 {
			if len(data[2:]) < ed25519.PublicKeySize {
				return nil, errors.New("deserializing public key failed: not enough length for Ed25519 key")
			}
			return ed25519.PublicKey(data[2:]), nil
		} else {
			return nil, errors.New("deserializing public key failed: unsupported EdDSA scheme")
		}

	case PK_P256_E, PK_P256_O:
		pub, err := ec.DecodePublicKey(data, elliptic.P256())
		if err != nil {
			return nil, errors.New("deserializing public key failed: decode P-256 public key error")
		}

		pk := &ec.PublicKey{
			Algorithm: ec.ECDSA,
			PublicKey: pub,
		}
		return pk, nil

	default:
		return nil, errors.New("deserializing public key failed: unrecognized algorithm label")
	}

}

// SerializePrivateKey serializes the input private key into byte array as the
// following format:
//              |--------------------|------------------|
//              | algorithm (1 byte) | private_key_data |
//              |--------------------|------------------|
//
// The private_key_data differs in algorithm:
//
// - ECDSA & SM2
//           |----------------|---------|--------------------|
//           | curve (1 byte) | d_bytes | encoded_public_key |
//           |----------------|---------|--------------------|
//   d_bytes is the byte sequence converted from the integer d in little
//   endian, with the byte length specified by the curve.
//   encoded_public_key is the public key data encoded in comopression mode.
//
// - EdDSA
//   Since only Ed25519 supported currently, it is just the 1 byte Ed25519
//   curve label followed by the byte sequence which could be handled as
//   private key in package ed25519.
//
// This function will panic if error occurs.
func SerializePrivateKey(pri PrivateKey) []byte {
	var buf bytes.Buffer
	switch t := pri.(type) {
	case *ec.PrivateKey:
		switch t.Algorithm {
		case ec.ECDSA:
			buf.WriteByte(byte(PK_ECDSA))
		case ec.SM2:
			buf.WriteByte(byte(PK_SM2))
		}
		label, err := GetCurveLabel(t.Curve)
		if err != nil {
			panic(err)
		}
		buf.WriteByte(label)
		size := (t.Params().BitSize + 7) >> 3
		dBytes := t.D.Bytes()
		for i := len(dBytes); i < size; i++ {
			buf.WriteByte(byte(0))
		}
		buf.Write(dBytes)
		buf.Write(ec.EncodePublicKey(&t.PublicKey, true))
	case ed25519.PrivateKey:
		buf.WriteByte(byte(PK_EDDSA))
		buf.WriteByte(byte(ED25519))
		buf.Write(t)
	default:
		panic("unkown private key type")
	}
	return buf.Bytes()
}

// DeserializePrivateKey parses the input byte array into private key.
func DeserializePrivateKey(data []byte) (pri PrivateKey, err error) {
	switch KeyType(data[0]) {
	case PK_ECDSA, PK_SM2:
		c, err1 := GetCurve(data[1])
		if err1 != nil {
			err = err1
			return
		}
		size := (c.Params().BitSize + 7) >> 3
		if len(data) < size*2+3 {
			err = errors.New("deserializing private key failed: not enough length")
			return
		}

		key := &ec.PrivateKey{
			Algorithm:  ec.ECDSA,
			PrivateKey: ec.ConstructPrivateKey(data[2:2+size], c),
		}

		p, err1 := ec.DecodePublicKey(data[2+size:], c)
		if err1 != nil {
			err = fmt.Errorf("deserializing private key failed: %s", err1)
			return
		}
		if key.X.Cmp(p.X) != 0 || key.Y.Cmp(p.Y) != 0 {
			err = errors.New("deserializing private key failed: unmatched private and public key")
			return
		}

		switch KeyType(data[0]) {
		case PK_ECDSA:
			key.Algorithm = ec.ECDSA
		case PK_SM2:
			key.Algorithm = ec.SM2
		}
		pri = key

	case PK_EDDSA:
		if data[1] == ED25519 {
			if len(data) < 2+ed25519.PrivateKeySize {
				err = errors.New("deserializing private key failed: not enough length for Ed25519 key")
				return
			}
			pri = ed25519.PrivateKey(data[2:])
		} else {
			err = errors.New("deserializing private key failed: unknown EdDSA curve type")
			return
		}
	}
	return
}

// ComparePublicKey checks whether the two public key are the same.
func ComparePublicKey(k0, k1 PublicKey) bool {
	if reflect.TypeOf(k0) != reflect.TypeOf(k1) {
		return false
	}

	switch v0 := k0.(type) {
	case *ec.PublicKey:
		v1 := k1.(*ec.PublicKey)
		if v0.Algorithm == v1.Algorithm && v0.Params().Name == v1.Params().Name && v0.X.Cmp(v1.X) == 0 {
			return true
		}

	case ed25519.PublicKey:
		v1 := k1.(ed25519.PublicKey)
		if bytes.Compare(v0, v1) == 0 {
			return true
		}
	}

	return false
}

// Parse ECDSA P-256 private key in WIF
func GetP256KeyPairFromWIF(wif []byte) (PrivateKey, error) {
	buf, err := base58.BitcoinEncoding.Decode(wif)
	if err != nil {
		return nil, err
	}
	bi, ok := new(big.Int).SetString(string(buf), 10)
	clearBytes(buf)
	if !ok || bi == nil {
		return nil, errors.New("parse WIF error, invalid base58 data")
	}
	buf = bi.Bytes()
	defer clearBytes(buf)
	pos := len(buf) - 4
	sum := sha256.Sum256(buf[:pos])
	sum = sha256.Sum256(sum[:])
	if !bytes.Equal(sum[:4], buf[pos:]) {
		return nil, errors.New("invalid WIF data, checksum failed")
	}
	pri := ec.ConstructPrivateKey(buf[1:pos-1], elliptic.P256())
	return &ec.PrivateKey{Algorithm: ec.ECDSA, PrivateKey: pri}, nil
}
