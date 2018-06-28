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

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains the Go wrapper for the constant-time, 64-bit assembly
// implementation of P256. The optimizations performed here are described in
// detail in:
// S.Gueron and V.Krasnov, "Fast prime field elliptic-curve cryptography with
//                          256-bit primes"
// http://link.springer.com/article/10.1007%2Fs13389-014-0090-x
// https://eprint.iacr.org/2013/816.pdf
// modify for sm2. Zhang Wei <d5c5ceb0@gmail.com>

// +build amd64

package sm2

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
	"sync"
)

type (
	p256Curve struct {
		*elliptic.CurveParams
		a []byte
	}

	p256Point struct {
		xyz [12]uint64
	}
)

var (
	p256            p256Curve
	p256Precomputed *[37][64 * 8]uint64
	precomputeOnce  sync.Once
)

func dump_ploy(str string, in []uint64) {

	fmt.Printf(str)
	for i := 3; i >= 0; i-- {
		fmt.Printf("%016x", in[i])
	}
	fmt.Printf("\n")

}

func dump_point(t1 []uint64) {
	fmt.Printf("set x ")
	for i := 3; i >= 0; i-- {
		fmt.Printf("%016x", t1[i])
	}
	fmt.Printf("\n")

	fmt.Printf("set y ")
	for i := 3; i >= 0; i-- {
		fmt.Printf("%016x", t1[i+4])
	}
	fmt.Printf("\n")

	fmt.Printf("set z ")
	for i := 3; i >= 0; i-- {
		fmt.Printf("%016x", t1[i+8])
	}
	fmt.Printf("\n")
}

func initP256() {
	// See FIPS 186-3, section D.2.3
	p256.CurveParams = &elliptic.CurveParams{Name: "sm2p256v1"}
	p256.P, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF", 16)
	p256.N, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123", 16)
	p256.B, _ = new(big.Int).SetString("28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93", 16)
	p256.Gx, _ = new(big.Int).SetString("32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7", 16)
	p256.Gy, _ = new(big.Int).SetString("BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0", 16)
	p256.BitSize = 256
	p256.a = []byte{0xFF, 0xFF, 0xFF, 0xFE, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFC}
}

func (curve p256Curve) Params() *elliptic.CurveParams {
	return curve.CurveParams
}

func (curve p256Curve) ABytes() []byte {
	return curve.a
}

// Functions implemented in p256_asm_amd64.s
// Montgomery multiplication modulo P256
func p256_sm2Mul(res, in1, in2 []uint64)

// Montgomery square modulo P256
func p256_sm2Sqr(res, in []uint64)

// Montgomery multiplication by 1
func p256_sm2FromMont(res, in []uint64)

// iff cond == 1  val <- -val
func p256_sm2NegCond(val []uint64, cond int)

// if cond == 0 res <- b; else res <- a
func p256_sm2MovCond(res, a, b []uint64, cond int)

// Endianness swap
func p256_sm2BigToLittle(res []uint64, in []byte)
func p256_sm2LittleToBig(res []byte, in []uint64)

// Constant time table access
func p256_sm2Select(point, table []uint64, idx int)
func p256_sm2SelectBase(point, table []uint64, idx int)

// Montgomery multiplication modulo Ord(G)
func p256_sm2OrdMul(res, in1, in2 []uint64)

// Montgomery square modulo Ord(G), repeated n times
func p256_sm2OrdSqr(res, in []uint64, n int)

// Point add with in2 being affine point
// If sign == 1 -> in2 = -in2
// If sel == 0 -> res = in1
// if zero == 0 -> res = in2
func p256_sm2PointAddAffineAsm(res, in1, in2 []uint64, sign, sel, zero int)

// Point add
func p256_sm2PointAddAsm(res, in1, in2 []uint64)

// Point double
func p256_sm2PointDoubleAsm(res, in []uint64)

func (curve p256Curve) Inverse(k *big.Int) *big.Int {
	if k.Sign() < 0 {
		// This should never happen.
		k = new(big.Int).Neg(k)
	}

	if k.Cmp(p256.N) >= 0 {
		// This should never happen.
		k = new(big.Int).Mod(k, p256.N)
	}

	// table will store precomputed powers of x. The four words at index
	// 4×i store x^(i+1).
	var table [4 * 15]uint64

	x := make([]uint64, 4)
	fromBig(x[:], k)
	// This code operates in the Montgomery domain where R = 2^256 mod n
	// and n is the order of the scalar field. (See initP256 for the
	// value.) Elements in the Montgomery domain take the form a×R and
	// multiplication of x and y in the calculates (x × y × R^-1) mod n. RR
	// is R×R mod n thus the Montgomery multiplication x and RR gives x×R,
	// i.e. converts x into the Montgomery domain.
	//RR := []uint64{0x83244c95be79eea2, 0x4699799c49bd6fa6, 0x2845b2392b6bec59, 0x66e12d94f3d95620}
	RR := []uint64{0x901192AF7C114F20, 0x3464504ADE6FA2FA, 0x620FC84C3AFFE0D4, 0x1EB5E412A22B3D3B}
	p256_sm2OrdMul(table[:4], x, RR)

	// Prepare the table, no need in constant time access, because the
	// power is not a secret. (Entry 0 is never used.)
	for i := 2; i < 16; i += 2 {
		p256_sm2OrdSqr(table[4*(i-1):], table[4*((i/2)-1):], 1)
		p256_sm2OrdMul(table[4*i:], table[4*(i-1):], table[:4])
	}

	x[0] = table[4*14+0] // f
	x[1] = table[4*14+1]
	x[2] = table[4*14+2]
	x[3] = table[4*14+3]

	p256_sm2OrdSqr(x, x, 4)
	p256_sm2OrdMul(x, x, table[4*14:4*14+4]) // ff
	t := make([]uint64, 4, 4)
	t[0] = x[0]
	t[1] = x[1]
	t[2] = x[2]
	t[3] = x[3]

	p256_sm2OrdSqr(x, x, 8)
	p256_sm2OrdMul(x, x, t) // ffff
	t[0] = x[0]
	t[1] = x[1]
	t[2] = x[2]
	t[3] = x[3]

	p256_sm2OrdSqr(x, x, 16)
	p256_sm2OrdMul(x, x, t) // ffffffff
	t[0] = x[0]
	t[1] = x[1]
	t[2] = x[2]
	t[3] = x[3]

	p256_sm2OrdSqr(x, x, 64) // ffffffff0000000000000000
	p256_sm2OrdMul(x, x, t)  // ffffffff00000000ffffffff
	p256_sm2OrdSqr(x, x, 32) // ffffffff00000000ffffffff00000000
	p256_sm2OrdMul(x, x, t)  // ffffffff00000000ffffffffffffffff

	// Remaining 32 windows
	expLo := [32]byte{0xb, 0xc, 0xe, 0x6, 0xf, 0xa, 0xa, 0xd, 0xa, 0x7, 0x1, 0x7, 0x9, 0xe, 0x8, 0x4, 0xf, 0x3, 0xb, 0x9, 0xc, 0xa, 0xc, 0x2, 0xf, 0xc, 0x6, 0x3, 0x2, 0x5, 0x4, 0xf}
	for i := 0; i < 32; i++ {
		p256_sm2OrdSqr(x, x, 4)
		p256_sm2OrdMul(x, x, table[4*(expLo[i]-1):])
	}

	// Multiplying by one in the Montgomery domain converts a Montgomery
	// value out of the domain.
	one := []uint64{1, 0, 0, 0}
	p256_sm2OrdMul(x, x, one)

	xOut := make([]byte, 32)
	p256_sm2LittleToBig(xOut, x)
	return new(big.Int).SetBytes(xOut)
}

// fromBig converts a *big.Int into a format used by this code.
func fromBig(out []uint64, big *big.Int) {
	for i := range out {
		out[i] = 0
	}

	for i, v := range big.Bits() {
		out[i] = uint64(v)
	}
}

// p256GetScalar endian-swaps the big-endian scalar value from in and writes it
// to out. If the scalar is equal or greater than the order of the group, it's
// reduced modulo that order.
func p256GetScalar(out []uint64, in []byte) {
	n := new(big.Int).SetBytes(in)

	if n.Cmp(p256.N) >= 0 {
		n.Mod(n, p256.N)
	}
	fromBig(out, n)
}

// p256_sm2Mul operates in a Montgomery domain with R = 2^256 mod p, where p is the
// underlying field of the curve. (See initP256 for the value.) Thus rr here is
// R×R mod p. See comment in Inverse about how this is used.
//var rr = []uint64{0x0000000000000003, 0xfffffffbffffffff, 0xfffffffffffffffe, 0x00000004fffffffd}
var rr = []uint64{0x0000000200000003, 0x00000002FFFFFFFF, 0x0000000100000001, 0x0000000400000002}

func maybeReduceModP(in *big.Int) *big.Int {
	if in.Cmp(p256.P) < 0 {
		return in
	}
	return new(big.Int).Mod(in, p256.P)
}

func (curve p256Curve) CombinedMult(bigX, bigY *big.Int, baseScalar, scalar []byte) (x, y *big.Int) {
	scalarReversed := make([]uint64, 4)
	var r1, r2 p256Point
	p256GetScalar(scalarReversed, baseScalar)
	r1.p256BaseMult(scalarReversed)

	p256GetScalar(scalarReversed, scalar)
	fromBig(r2.xyz[0:4], maybeReduceModP(bigX))
	fromBig(r2.xyz[4:8], maybeReduceModP(bigY))
	p256_sm2Mul(r2.xyz[0:4], r2.xyz[0:4], rr[:])
	p256_sm2Mul(r2.xyz[4:8], r2.xyz[4:8], rr[:])

	// This sets r2's Z value to 1, in the Montgomery domain.
	//r2.xyz[8] = 0x0000000000000001
	//r2.xyz[9] = 0xffffffff00000000
	//r2.xyz[10] = 0xffffffffffffffff
	//r2.xyz[11] = 0x00000000fffffffe
	r2.xyz[8] = 0x0000000000000001
	r2.xyz[9] = 0x00000000FFFFFFFF
	r2.xyz[10] = 0x0000000000000000
	r2.xyz[11] = 0x0000000100000000

	r2.p256ScalarMult(scalarReversed)
	p256_sm2PointAddAsm(r1.xyz[:], r1.xyz[:], r2.xyz[:])
	return r1.p256PointToAffine()
}

func (curve p256Curve) ScalarBaseMult(scalar []byte) (x, y *big.Int) {
	scalarReversed := make([]uint64, 4)
	p256GetScalar(scalarReversed, scalar)
	var r p256Point
	r.p256BaseMult(scalarReversed)

	return r.p256PointToAffine()
}

func (curve p256Curve) ScalarMult(bigX, bigY *big.Int, scalar []byte) (x, y *big.Int) {
	scalarReversed := make([]uint64, 4)
	p256GetScalar(scalarReversed, scalar)

	var r p256Point
	fromBig(r.xyz[0:4], maybeReduceModP(bigX))
	fromBig(r.xyz[4:8], maybeReduceModP(bigY))
	p256_sm2Mul(r.xyz[0:4], r.xyz[0:4], rr[:])
	p256_sm2Mul(r.xyz[4:8], r.xyz[4:8], rr[:])
	// This sets r2's Z value to 1, in the Montgomery domain.
	//r.xyz[8] = 0x0000000000000001
	//r.xyz[9] = 0xffffffff00000000
	//r.xyz[10] = 0xffffffffffffffff
	//r.xyz[11] = 0x00000000fffffffe
	r.xyz[8] = 0x0000000000000001
	r.xyz[9] = 0x00000000FFFFFFFF
	r.xyz[10] = 0x0000000000000000
	r.xyz[11] = 0x0000000100000000

	r.p256ScalarMult(scalarReversed)
	return r.p256PointToAffine()
}

func (p *p256Point) p256PointToAffine() (x, y *big.Int) {
	zInv := make([]uint64, 4)
	zInvSq := make([]uint64, 4)
	p256Inverse(zInv, p.xyz[8:12])
	p256_sm2Sqr(zInvSq, zInv)
	p256_sm2Mul(zInv, zInv, zInvSq)

	p256_sm2Mul(zInvSq, p.xyz[0:4], zInvSq)
	p256_sm2Mul(zInv, p.xyz[4:8], zInv)

	p256_sm2FromMont(zInvSq, zInvSq)
	p256_sm2FromMont(zInv, zInv)

	xOut := make([]byte, 32)
	yOut := make([]byte, 32)
	p256_sm2LittleToBig(xOut, zInvSq)
	p256_sm2LittleToBig(yOut, zInv)

	return new(big.Int).SetBytes(xOut), new(big.Int).SetBytes(yOut)
}

func p256Inverse(out, in []uint64) {
	var stack [6 * 4]uint64
	p2 := stack[4*0 : 4*0+4]
	p4 := stack[4*1 : 4*1+4]
	p8 := stack[4*2 : 4*2+4]
	p16 := stack[4*3 : 4*3+4]
	p32 := stack[4*4 : 4*4+4]

	p256_sm2Sqr(out, in)
	p256_sm2Mul(p2, out, in) // 3*p

	p256_sm2Sqr(out, p2)
	p256_sm2Sqr(out, out)
	p256_sm2Mul(p4, out, p2) // f*p

	p256_sm2Sqr(out, p4)
	p256_sm2Sqr(out, out)
	p256_sm2Sqr(out, out)
	p256_sm2Sqr(out, out)
	p256_sm2Mul(p8, out, p4) // ff*p

	p256_sm2Sqr(out, p8)

	for i := 0; i < 7; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(p16, out, p8) // ffff*p

	p256_sm2Sqr(out, p16)
	for i := 0; i < 15; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(p32, out, p16) // ffffffff*p

	p256_sm2Sqr(out, p16)
	for i := 0; i < 7; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p8)
	for i := 0; i < 4; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p4)
	p256_sm2Sqr(out, out)
	p256_sm2Sqr(out, out)
	p256_sm2Mul(out, out, p2)

	p256_sm2Sqr(out, out)
	p256_sm2Mul(out, out, in)
	p256_sm2Sqr(out, out) //fffffffe*p

	for i := 0; i < 32; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p32) //fffffffeffffffff*p

	for i := 0; i < 32; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p32) //fffffffeffffffffffffffff*p

	for i := 0; i < 32; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p32) //fffffffe ffffffff ffffffff ffffffff*p

	for i := 0; i < 32; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p32) //fffffffe ffffffff ffffffff ffffffff ffffffff*p

	for i := 0; i < 32*2; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p32) //fffffffe ffffffff ffffffff ffffffff ffffffff 00000000 ffffffff*p

	for i := 0; i < 16; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p16)

	for i := 0; i < 8; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p8)

	for i := 0; i < 4; i++ {
		p256_sm2Sqr(out, out)
	}
	p256_sm2Mul(out, out, p4)

	p256_sm2Sqr(out, out)
	p256_sm2Sqr(out, out)
	p256_sm2Mul(out, out, p2)

	p256_sm2Sqr(out, out)
	p256_sm2Sqr(out, out)
	p256_sm2Mul(out, out, in)
}

func (p *p256Point) p256StorePoint(r *[16 * 4 * 3]uint64, index int) {
	copy(r[index*12:], p.xyz[:])
}

func boothW5(in uint) (int, int) {
	var s uint = ^((in >> 5) - 1)
	var d uint = (1 << 6) - in - 1
	d = (d & s) | (in & (^s))
	d = (d >> 1) + (d & 1)
	return int(d), int(s & 1)
}

func boothW7(in uint) (int, int) {
	var s uint = ^((in >> 7) - 1)
	var d uint = (1 << 8) - in - 1
	d = (d & s) | (in & (^s))
	d = (d >> 1) + (d & 1)
	return int(d), int(s & 1)
}

func initTable() {
	p256Precomputed = new([37][64 * 8]uint64)

	basePoint := []uint64{
		//0x79e730d418a9143c, 0x75ba95fc5fedb601, 0x79fb732b77622510, 0x18905f76a53755c6,
		0x61328990F418029E, 0x3E7981EDDCA6C050, 0xD6A1ED99AC24C3C3, 0x91167A5EE1C13B05,
		//0xddf25357ce95560a, 0x8b4ab8e4ba19e45c, 0xd2e88688dd21f325, 0x8571ff1825885d85,
		0xC1354E593C2D0DDD, 0xC1F5E5788D3295FA, 0x8D4CFB066E2A48F8, 0x63CD65D481D735BD,
		//0x0000000000000001, 0xffffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe,
		0x0000000000000001, 0x00000000FFFFFFFF, 0x0000000000000000, 0x0000000100000000,
	}
	t1 := make([]uint64, 12)
	t2 := make([]uint64, 12)
	copy(t2, basePoint)

	zInv := make([]uint64, 4)
	zInvSq := make([]uint64, 4)
	for j := 0; j < 64; j++ {
		copy(t1, t2)
		for i := 0; i < 37; i++ {
			// The window size is 7 so we need to double 7 times.
			if i != 0 {
				for k := 0; k < 7; k++ {
					p256_sm2PointDoubleAsm(t1, t1)
				}
			}

			// Convert the point to affine form. (Its values are
			// still in Montgomery form however.)

			p256Inverse(zInv, t1[8:12])

			p256_sm2Sqr(zInvSq, zInv)
			p256_sm2Mul(zInv, zInv, zInvSq)

			p256_sm2Mul(t1[:4], t1[:4], zInvSq)
			p256_sm2Mul(t1[4:8], t1[4:8], zInv)

			copy(t1[8:12], basePoint[8:12])
			// Update the table entry
			copy(p256Precomputed[i][j*8:], t1[:8])
		}
		if j == 0 {
			p256_sm2PointDoubleAsm(t2, basePoint)
		} else {
			p256_sm2PointAddAsm(t2, t2, basePoint)
		}
	}
}

func (p *p256Point) p256BaseMult(scalar []uint64) {
	precomputeOnce.Do(initTable)

	wvalue := (scalar[0] << 1) & 0xff
	sel, sign := boothW7(uint(wvalue))
	p256_sm2SelectBase(p.xyz[0:8], p256Precomputed[0][0:], sel)
	p256_sm2NegCond(p.xyz[4:8], sign)

	// (This is one, in the Montgomery domain.)
	//p.xyz[8] = 0x0000000000000001
	//p.xyz[9] = 0xffffffff00000000
	//p.xyz[10] = 0xffffffffffffffff
	//p.xyz[11] = 0x00000000fffffffe
	p.xyz[8] = 0x0000000000000001
	p.xyz[9] = 0x00000000FFFFFFFF
	p.xyz[10] = 0x0000000000000000
	p.xyz[11] = 0x0000000100000000

	var t0 p256Point
	// (This is one, in the Montgomery domain.)
	//t0.xyz[8] = 0x0000000000000001
	//t0.xyz[9] = 0xffffffff00000000
	//t0.xyz[10] = 0xffffffffffffffff
	//t0.xyz[11] = 0x00000000fffffffe
	t0.xyz[8] = 0x0000000000000001
	t0.xyz[9] = 0x00000000FFFFFFFF
	t0.xyz[10] = 0x0000000000000000
	t0.xyz[11] = 0x0000000100000000

	index := uint(6)
	zero := sel

	for i := 1; i < 37; i++ {
		if index < 192 {
			wvalue = ((scalar[index/64] >> (index % 64)) + (scalar[index/64+1] << (64 - (index % 64)))) & 0xff
		} else {
			wvalue = (scalar[index/64] >> (index % 64)) & 0xff
		}
		index += 7
		sel, sign = boothW7(uint(wvalue))
		p256_sm2SelectBase(t0.xyz[0:8], p256Precomputed[i][0:], sel)
		p256_sm2PointAddAffineAsm(p.xyz[0:12], p.xyz[0:12], t0.xyz[0:8], sign, sel, zero)
		zero |= sel
	}
}

func (p *p256Point) p256ScalarMult(scalar []uint64) {
	// precomp is a table of precomputed points that stores powers of p
	// from p^1 to p^16.
	var precomp [16 * 4 * 3]uint64
	var t0, t1, t2, t3 p256Point

	// Prepare the table
	p.p256StorePoint(&precomp, 0) // 1

	p256_sm2PointDoubleAsm(t0.xyz[:], p.xyz[:])
	p256_sm2PointDoubleAsm(t1.xyz[:], t0.xyz[:])
	p256_sm2PointDoubleAsm(t2.xyz[:], t1.xyz[:])
	p256_sm2PointDoubleAsm(t3.xyz[:], t2.xyz[:])
	t0.p256StorePoint(&precomp, 1)  // 2
	t1.p256StorePoint(&precomp, 3)  // 4
	t2.p256StorePoint(&precomp, 7)  // 8
	t3.p256StorePoint(&precomp, 15) // 16

	p256_sm2PointAddAsm(t0.xyz[:], t0.xyz[:], p.xyz[:])
	p256_sm2PointAddAsm(t1.xyz[:], t1.xyz[:], p.xyz[:])
	p256_sm2PointAddAsm(t2.xyz[:], t2.xyz[:], p.xyz[:])
	t0.p256StorePoint(&precomp, 2) // 3
	t1.p256StorePoint(&precomp, 4) // 5
	t2.p256StorePoint(&precomp, 8) // 9

	p256_sm2PointDoubleAsm(t0.xyz[:], t0.xyz[:])
	p256_sm2PointDoubleAsm(t1.xyz[:], t1.xyz[:])
	t0.p256StorePoint(&precomp, 5) // 6
	t1.p256StorePoint(&precomp, 9) // 10

	p256_sm2PointAddAsm(t2.xyz[:], t0.xyz[:], p.xyz[:])
	p256_sm2PointAddAsm(t1.xyz[:], t1.xyz[:], p.xyz[:])
	t2.p256StorePoint(&precomp, 6)  // 7
	t1.p256StorePoint(&precomp, 10) // 11

	p256_sm2PointDoubleAsm(t0.xyz[:], t0.xyz[:])
	p256_sm2PointDoubleAsm(t2.xyz[:], t2.xyz[:])
	t0.p256StorePoint(&precomp, 11) // 12
	t2.p256StorePoint(&precomp, 13) // 14

	p256_sm2PointAddAsm(t0.xyz[:], t0.xyz[:], p.xyz[:])
	p256_sm2PointAddAsm(t2.xyz[:], t2.xyz[:], p.xyz[:])
	t0.p256StorePoint(&precomp, 12) // 13
	t2.p256StorePoint(&precomp, 14) // 15

	// Start scanning the window from top bit
	index := uint(254)
	var sel, sign int

	wvalue := (scalar[index/64] >> (index % 64)) & 0x3f
	sel, _ = boothW5(uint(wvalue))

	p256_sm2Select(p.xyz[0:12], precomp[0:], sel)
	zero := sel

	for index > 4 {
		index -= 5
		p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])
		p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])
		p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])
		p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])
		p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])

		if index < 192 {
			wvalue = ((scalar[index/64] >> (index % 64)) + (scalar[index/64+1] << (64 - (index % 64)))) & 0x3f
		} else {
			wvalue = (scalar[index/64] >> (index % 64)) & 0x3f
		}

		sel, sign = boothW5(uint(wvalue))

		p256_sm2Select(t0.xyz[0:], precomp[0:], sel)
		p256_sm2NegCond(t0.xyz[4:8], sign)
		p256_sm2PointAddAsm(t1.xyz[:], p.xyz[:], t0.xyz[:])
		p256_sm2MovCond(t1.xyz[0:12], t1.xyz[0:12], p.xyz[0:12], sel)
		p256_sm2MovCond(p.xyz[0:12], t1.xyz[0:12], t0.xyz[0:12], zero)
		zero |= sel
	}

	p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])
	p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])
	p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])
	p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])
	p256_sm2PointDoubleAsm(p.xyz[:], p.xyz[:])

	wvalue = (scalar[0] << 1) & 0x3f
	sel, sign = boothW5(uint(wvalue))

	p256_sm2Select(t0.xyz[0:], precomp[0:], sel)
	p256_sm2NegCond(t0.xyz[4:8], sign)
	p256_sm2PointAddAsm(t1.xyz[:], p.xyz[:], t0.xyz[:])
	p256_sm2MovCond(t1.xyz[0:12], t1.xyz[0:12], p.xyz[0:12], sel)
	p256_sm2MovCond(p.xyz[0:12], t1.xyz[0:12], t0.xyz[0:12], zero)
}
