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
	"crypto"
	"errors"
	"fmt"
	"hash"
	"strings"

	"github.com/ontio/ontology-crypto/sm3"

	// the following blank imports ensures these packages are linked
	_ "crypto/sha256"
	_ "crypto/sha512"
	_ "golang.org/x/crypto/ripemd160"
	_ "golang.org/x/crypto/sha3"
)

type SignatureScheme byte

const (
	SHA224withECDSA SignatureScheme = iota
	SHA256withECDSA
	SHA384withECDSA
	SHA512withECDSA
	SHA3_224withECDSA
	SHA3_256withECDSA
	SHA3_384withECDSA
	SHA3_512withECDSA
	RIPEMD160withECDSA

	SM3withSM2

	SHA512withEDDSA
)

var names []string = []string{
	"SHA224withECDSA",
	"SHA256withECDSA",
	"SHA384withECDSA",
	"SHA512withECDSA",
	"SHA3-224withECDSA",
	"SHA3-256withECDSA",
	"SHA3-384withECDSA",
	"SHA3-512withECDSA",
	"RIPEMD160withECDSA",
	"SM3withSM2",
	"SHA512withEdDSA",
}

func (s SignatureScheme) Name() string {
	if int(s) >= len(names) {
		panic(fmt.Sprintf("unknown scheme value %v", s))
	}
	return names[s]
}

func GetScheme(name string) (SignatureScheme, error) {
	for i, v := range names {
		if strings.ToUpper(v) == strings.ToUpper(name) {
			return SignatureScheme(i), nil
		}
	}

	return 0, errors.New("unknown signature scheme " + name)
}

func GetHash(scheme SignatureScheme) hash.Hash {
	switch scheme {
	case SHA224withECDSA:
		return crypto.SHA224.New()
	case SHA256withECDSA:
		return crypto.SHA256.New()
	case SHA384withECDSA:
		return crypto.SHA384.New()
	case SHA512withECDSA:
		return crypto.SHA512.New()
	case SHA3_224withECDSA:
		return crypto.SHA3_224.New()
	case SHA3_256withECDSA:
		return crypto.SHA3_256.New()
	case SHA3_384withECDSA:
		return crypto.SHA3_384.New()
	case SHA3_512withECDSA:
		return crypto.SHA3_512.New()
	case RIPEMD160withECDSA:
		return crypto.RIPEMD160.New()
	case SM3withSM2:
		return sm3.New()
	case SHA512withEDDSA:
		return crypto.SHA512.New()
	}
	return nil
}
