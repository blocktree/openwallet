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
	"crypto/elliptic"
	"errors"
	"strings"

	"github.com/ontio/ontology-crypto/sm2"
)

const (
	// ECDSA curve label
	P224 byte = 1
	P256 byte = 2
	P384 byte = 3
	P521 byte = 4

	// SM2 curve label
	SM2P256V1 byte = 20

	// ED25519 curve label
	ED25519 byte = 25
)

func GetCurveLabel(c elliptic.Curve) (byte, error) {
	return GetNamedCurveLabel(c.Params().Name)
}

func GetCurve(label byte) (elliptic.Curve, error) {
	switch label {
	case P224:
		return elliptic.P224(), nil
	case P256:
		return elliptic.P256(), nil
	case P384:
		return elliptic.P384(), nil
	case P521:
		return elliptic.P521(), nil
	case SM2P256V1:
		return sm2.SM2P256V1(), nil
	default:
		return nil, errors.New("unknown elliptic curve")
	}

}

func GetNamedCurve(name string) (elliptic.Curve, error) {
	label, err := GetNamedCurveLabel(name)
	if err != nil {
		return nil, err
	}
	return GetCurve(label)
}

func GetNamedCurveLabel(name string) (byte, error) {
	switch strings.ToUpper(name) {
	case strings.ToUpper(elliptic.P224().Params().Name):
		return P224, nil
	case strings.ToUpper(elliptic.P256().Params().Name):
		return P256, nil
	case strings.ToUpper(elliptic.P384().Params().Name):
		return P384, nil
	case strings.ToUpper(elliptic.P521().Params().Name):
		return P521, nil
	case strings.ToUpper(sm2.SM2P256V1().Params().Name):
		return SM2P256V1, nil
	default:
		return 0, errors.New("unsupported elliptic curve")
	}
}
