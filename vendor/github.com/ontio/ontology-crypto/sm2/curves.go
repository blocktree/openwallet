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

package sm2

import (
	"crypto/elliptic"
	"sync"
)

var initonce sync.Once

// SM2Curve is the curve interface used in sm2 algorithm.
// It extends elliptic.Curve by adding a function ABytes().
type SM2Curve interface {
	elliptic.Curve

	// ABytes returns the little endian byte sequence of parameter A.
	ABytes() []byte
}

// SM2P256V1 returns the sm2p256v1 curve.
func SM2P256V1() elliptic.Curve {
	initonce.Do(initP256)
	return p256
}
