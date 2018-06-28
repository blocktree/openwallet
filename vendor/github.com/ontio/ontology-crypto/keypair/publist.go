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

package keypair

import (
	"bytes"
	"sort"

	"github.com/ontio/ontology-crypto/ec"

	"golang.org/x/crypto/ed25519"
)

// SorPublicKeys sorts the input PublicKey slice and return it.
// The sorting rules is as follows:
//    1. if keys have different types, then sorted by the KeyType value.
//    2. else,
//       2.1. ECDSA or SM2:
//           2.1.1. if on different curves, then sorted by the curve label.
//           2.1.2. else if x values are different, then sorted by x.
//           2.1.3. else sorted by y.
//       2.2. EdDSA: sorted by the byte sequence directly.
func SortPublicKeys(list []PublicKey) []PublicKey {
	pl := publicKeyList(list)
	sort.Sort(pl)
	return pl
}

type publicKeyList []PublicKey

func (this publicKeyList) Len() int {
	return len(this)
}

func (this publicKeyList) Less(i, j int) bool {
	a, b := this[i], this[j]
	ta := GetKeyType(a)
	tb := GetKeyType(b)
	if ta != tb {
		return ta < tb
	}

	switch ta {
	case PK_ECDSA, PK_SM2:
		va := a.(*ec.PublicKey)
		vb := b.(*ec.PublicKey)
		ca, err := GetCurveLabel(va)
		if err != nil {
			panic(err)
		}
		cb, err := GetCurveLabel(vb)
		if err != nil {
			panic(err)
		}
		if ca != cb {
			return ca < cb
		}
		cmp := va.X.Cmp(vb.X)
		if cmp != 0 {
			return cmp < 0
		}
		cmp = va.Y.Cmp(vb.Y)
		return cmp < 0
	case PK_EDDSA:
		va := a.(ed25519.PublicKey)
		vb := b.(ed25519.PublicKey)
		return bytes.Compare(va, vb) < 0
	default:
		panic("error key type")
	}
	return true
}

func (this publicKeyList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

// FindKey finds the specified public key in the list and returns its index
// or -1 if not found.
func FindKey(list []PublicKey, key PublicKey) int {
	for i, v := range list {
		if ComparePublicKey(v, key) {
			return i
		}
	}
	return -1
}

// PublicList is a container for serialized public keys.
// It implements the interface sort.Interface.
type PublicList [][]byte

func (l PublicList) Len() int {
	return len(l)
}

func (l PublicList) Less(i, j int) bool {
	return bytes.Compare(l[i], l[j]) < 0
}

func (l PublicList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// ConvertToPublicList converts the public keys to a PublicList.
func NewPublicList(keys []PublicKey) PublicList {
	res := make(PublicList, 0, len(keys))
	for _, k := range keys {
		res = append(res, SerializePublicKey(k))
	}

	return res
}
