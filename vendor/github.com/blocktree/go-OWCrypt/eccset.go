package owcrypt

// #cgo CFLAGS: -I./csource/bignum
// #cgo CFLAGS: -I./csource/hash_drv
// #cgo CFLAGS: -I./csource/ecc_drv
// #cgo CFLAGS: -I./csource/Crypto
// #include <stdio.h>
// #include <stdlib.h>
// #include "type.h"
// #include "bignum.c"
// #include "bignum.h"
// #include "bigrand.c"
// #include "bigrand.h"
// #include "montgamery.c"
// #include "montgamery.h"
// #include "ecc_drv.c"
// #include "ecc_drv.h"
// #include "ECDSA.c"
// #include "ECDSA.h"
// #include "ED25519.c"
// #include "ED25519.h"
// #include "secp256k1.c"
// #include "secp256k1.h"
// #include "secp256r1.c"
// #include "secp256r1.h"
// #include "sm2.c"
// #include "sm2.h"
// #include "blake2b.c"
// #include "blake2b.h"
// #include "blake2s.c"
// #include "blake2s.h"
// #include "hmac.c"
// #include "hmac.h"
// #include "md4.c"
// #include "md4.h"
// #include "md5.c"
// #include "md5.h"
// #include "ripemd160.c"
// #include "ripemd160.h"
// #include "sha1.c"
// #include "sha1.h"
// #include "sm3.c"
// #include "sm3.h"
// #include "sha256.c"
// #include "sha256.h"
// #include "sha512.c"
// #include "sha512.h"
// #include "ecc_set.c"
// #include "ecc_set.h"
// #include "hash_set.c"
// #include "hash_set.h"
import "C"
import "unsafe"

const (
	HASH_ALG_SHA1          = uint32(0xA0000000)
	HASH_ALG_SHA256        = uint32(0xA0000001)
	HASH_ALG_SHA512        = uint32(0xA0000002)
	HASH_ALG_MD4           = uint32(0xA0000003)
	HASH_ALG_MD5           = uint32(0xA0000004)
	HASH_ALG_RIPEMD160     = uint32(0xA0000005)
	HASH_ALG_BLAKE2B       = uint32(0xA0000006)
	HASH_ALG_BLAKE2S       = uint32(0xA0000007)
	HASH_ALG_SM3           = uint32(0xA0000008)
	HASh_ALG_DOUBLE_SHA256 = uint32(0xA000000A)
	HASH_ALG_HASH160       = uint32(0xA000000B)

	ECC_CURVE_SECP256K1    = uint32(0xECC00000)
	ECC_CURVE_SECP256R1    = uint32(0xECC00001)
	ECC_CURVE_PRIMEV1      = ECC_CURVE_SECP256R1
	ECC_CURVE_NIST_P256    = ECC_CURVE_SECP256R1
	ECC_CURVE_SM2_STANDARD = uint32(0xECC00002)
	ECC_CURVE_ED25519      = uint32(0xECC00003)

	SUCCESS            = uint16(0x0001)
	FAILURE            = uint16(0x0000)
	ECC_PRIKEY_ILLEGAL = uint16(0xE000)
	ECC_PUBKEY_ILLEGAL = uint16(0xE001)
	ECC_WRONG_TYPE     = uint16(0xE002)
	ECC_MISS_ID        = uint16(0xE003)
)

func GenPubkey(prikey []byte, typeChoose uint32) (pubkey []byte, ret uint16) {
	var keylen uint16
	if typeChoose == ECC_CURVE_ED25519 {
		keylen = 32
	} else {
		keylen = 64
	}
	pubkey = make([]byte, keylen)
	pri := (*C.uchar)(unsafe.Pointer(&prikey[0]))
	pub := (*C.uchar)(unsafe.Pointer(&pubkey[0]))

	ret = uint16(C.ECC_genPubkey(pri, pub, C.uint(typeChoose)))

	return pubkey[:], ret
}

func Signature(prikey []byte, ID []byte, IDlen uint16, message []byte, message_len uint16, typeChoose uint32) (signature []byte, ret uint16) {
	signature = make([]byte, 64)
	pri := (*C.uchar)(unsafe.Pointer(&prikey[0]))
	id := (*C.uchar)(unsafe.Pointer(&ID[0]))
	msg := (*C.uchar)(unsafe.Pointer(&message[0]))
	sig := (*C.uchar)(unsafe.Pointer(&signature[0]))

	ret = uint16(C.ECC_sign(pri, id, C.ushort(IDlen), msg, C.ushort(message_len), sig, C.uint(typeChoose)))

	return signature, ret
}

func Verify(pubkey []byte, ID []byte, IDlen uint16, message []byte, message_len uint16, signature []byte, typeChoose uint32) uint16 {
	pub := (*C.uchar)(unsafe.Pointer(&pubkey[0]))
	id := (*C.uchar)(unsafe.Pointer(&ID[0]))
	msg := (*C.uchar)(unsafe.Pointer(&message[0]))
	sig := (*C.uchar)(unsafe.Pointer(&signature[0]))
	ret := C.ushort(0)

	ret = C.ECC_verify(pub, id, C.ushort(IDlen), msg, C.ushort(message_len), sig, C.uint(typeChoose))

	return uint16(ret)
}

func Encryption(pubkey []byte, plain []byte, typeChoose uint32) (cipher []byte, ret uint16) {
	plain_len := len(plain)
	cipher_len := uint16(plain_len + 97)
	cipher = make([]byte, cipher_len)
	pub := (*C.uchar)(unsafe.Pointer(&pubkey[0]))
	m := (*C.uchar)(unsafe.Pointer(&plain[0]))
	c := (*C.uchar)(unsafe.Pointer(&cipher[0]))

	ret = uint16(C.ECC_enc(pub, m, C.ushort(plain_len), c, (*C.ushort)(unsafe.Pointer(&cipher_len)), C.uint(typeChoose)))
	return cipher, ret
}

func Decryption(prikey []byte, cipher []byte, typeChoose uint32) (plain []byte, ret uint16) {
	cipher_len := len(cipher)
	plain_len := cipher_len - 97
	plain = make([]byte, plain_len)
	pri := (*C.uchar)(unsafe.Pointer(&prikey[0]))
	c := (*C.uchar)(unsafe.Pointer(&cipher[0]))
	m := (*C.uchar)(unsafe.Pointer(&plain[0]))

	ret = uint16(C.ECC_dec(pri, c, C.ushort(cipher_len), m, (*C.ushort)(unsafe.Pointer(&plain_len)), C.uint(typeChoose)))

	return plain, ret
}

///////////////////////////////////////////////////////////密钥协商////////////////////////////////////////////////////////
func KeyAgreement_initiator_step1(typeChoose uint32) (tmpPrikeyInitiator, tmpPubkeyInitiator []byte) {
	tmpPrikeyInitiator = make([]byte, 32)
	tmpPubkeyInitiator = make([]byte, 64)
	tmpPri := (*C.uchar)(unsafe.Pointer(&tmpPrikeyInitiator[0]))
	tmpPub := (*C.uchar)(unsafe.Pointer(&tmpPubkeyInitiator[0]))

	C.ECC_key_exchange_initiator_step1(tmpPri, tmpPub, C.uint(typeChoose))
	return tmpPrikeyInitiator, tmpPubkeyInitiator
}

func KeyAgreement_initiator_step2(IDinitiator []byte,
	IDinitiator_len uint16,
	IDresponder []byte,
	IDresponder_len uint16,
	prikeyInitiator []byte,
	pubkeyInitiator []byte,
	pubkeyResponder []byte,
	tmpPrikeyInitiator []byte,
	tmpPubkeyInitiator []byte,
	tmpPubkeyResponder []byte,
	Sin []byte,
	keylen uint16,
	typeChoose uint32) (key, Sout []byte, ret uint16) {
	//------------------------------------------------------------//
	key = make([]byte, keylen)
	Sout = make([]byte, 32)
	idInit := (*C.uchar)(unsafe.Pointer(&IDinitiator[0]))
	idResp := (*C.uchar)(unsafe.Pointer(&IDresponder[0]))
	priInit := (*C.uchar)(unsafe.Pointer(&prikeyInitiator[0]))
	pubInit := (*C.uchar)(unsafe.Pointer(&pubkeyInitiator[0]))
	pubResp := (*C.uchar)(unsafe.Pointer(&pubkeyResponder[0]))
	tmpPriInit := (*C.uchar)(unsafe.Pointer(&tmpPrikeyInitiator[0]))
	tmpPubInit := (*C.uchar)(unsafe.Pointer(&tmpPubkeyInitiator[0]))
	tmpPubResp := (*C.uchar)(unsafe.Pointer(&tmpPubkeyResponder[0]))
	sIn := (*C.uchar)(unsafe.Pointer(&Sin[0]))
	sOut := (*C.uchar)(unsafe.Pointer(&Sout[0]))
	result := (*C.uchar)(unsafe.Pointer(&key[0]))

	ret = uint16(C.ECC_key_exchange_initiator_step2(idInit,
		C.ushort(IDinitiator_len),
		idResp,
		C.ushort(IDresponder_len),
		priInit,
		pubInit,
		pubResp,
		tmpPriInit,
		tmpPubInit,
		tmpPubResp,
		sIn,
		sOut,
		C.ushort(keylen),
		result,
		C.uint(typeChoose)))

	return key, Sout, ret
}

func KeyAgreement_responder_step1(IDinitiator []byte,
	IDinitiator_len uint16,
	IDresponder []byte,
	IDresponder_len uint16,
	prikeyResponder []byte,
	pubkeyResponder []byte,
	pubkeyInitiator []byte,
	tmpPubkeyInitiator []byte,
	keylen uint16,
	typeChoose uint32) (key, tmpPubkeyResponder, Sinner, Souter []byte, ret uint16) {
	//------------------------------------------------------------//
	key = make([]byte, keylen)
	tmpPubkeyResponder = make([]byte, 64)
	Sinner = make([]byte, 32)
	Souter = make([]byte, 32)
	idInit := (*C.uchar)(unsafe.Pointer(&IDinitiator[0]))
	idResp := (*C.uchar)(unsafe.Pointer(&IDresponder[0]))
	priResp := (*C.uchar)(unsafe.Pointer(&prikeyResponder[0]))
	pubResp := (*C.uchar)(unsafe.Pointer(&pubkeyResponder[0]))
	pubInit := (*C.uchar)(unsafe.Pointer(&pubkeyInitiator[0]))
	tmpPubResp := (*C.uchar)(unsafe.Pointer(&tmpPubkeyResponder[0]))
	tmpPubInit := (*C.uchar)(unsafe.Pointer(&tmpPubkeyInitiator[0]))
	sInner := (*C.uchar)(unsafe.Pointer(&Sinner[0]))
	sOuter := (*C.uchar)(unsafe.Pointer(&Souter[0]))
	result := (*C.uchar)(unsafe.Pointer(&key[0]))

	ret = uint16(C.ECC_key_exchange_responder_step1(idInit,
		C.ushort(IDinitiator_len),
		idResp,
		C.ushort(IDresponder_len),
		priResp,
		pubResp,
		pubInit,
		tmpPubResp,
		tmpPubInit,
		sInner,
		sOuter,
		C.ushort(keylen),
		result,
		C.uint(typeChoose)))

	return key, tmpPubkeyResponder, Sinner, Souter, ret
}

func KeyAgreement_responder_step2(Sinitiator []byte, Sresponder []byte, typeChoose uint32) uint16 {
	sInit := (*C.uchar)(unsafe.Pointer(&Sinitiator[0]))
	sResp := (*C.uchar)(unsafe.Pointer(&Sresponder[0]))
	ret := C.ushort(0)

	ret = C.ECC_key_exchange_responder_step2(sInit, sResp, C.uint(typeChoose))

	return uint16(ret)
}

func Point_mulBaseG(scalar []byte, typeChoose uint32) []byte {
	var size uint16
	if typeChoose == ECC_CURVE_ED25519 {
		size = 32
	} else {
		size = 64
	}
	ret := make([]byte, size)

	pointOut := (*C.uchar)(unsafe.Pointer(&ret[0]))
	k := (*C.uchar)(unsafe.Pointer(&scalar[0]))

	C.ECC_point_mul_baseG(k, pointOut, C.uint(typeChoose))
	if typeChoose == ECC_CURVE_ED25519 {
		return ret
	}
	return PointCompress(ret[:], typeChoose)

}

//all ed25519 data is in little-endian
func Point_mulBaseG_add(pointin, scalar []byte, typeChoose uint32) (point []byte, isinfinity bool) {
	var size uint16
	if typeChoose == ECC_CURVE_ED25519 {
		size = 32
	} else {
		size = 64
	}
	ret := make([]byte, size)

	pointOut := (*C.uchar)(unsafe.Pointer(&ret[0]))
	pointIn := (*C.uchar)(unsafe.Pointer(&pointin[0]))
	k := (*C.uchar)(unsafe.Pointer(&scalar[0]))

	if C.ushort(C.ECC_point_mul_baseG_add(pointIn, k, pointOut, C.uint(typeChoose))) == 1 {
		return nil, true
	}

	return ret, false
}

func GetCurveOrder(typeChoose uint32) []byte {
	ret := [32]byte{}
	order := (*C.uchar)(unsafe.Pointer(&ret[0]))
	C.ECC_get_curve_order(order, C.uint(typeChoose))
	return ret[:]
}

func PointCompress(point []byte, typeChoose uint32) []byte {
	if typeChoose == ECC_CURVE_ED25519 {
		return point
	}
	ret := [33]byte{}
	pin := (*C.uchar)(unsafe.Pointer(&point[0]))
	pout := (*C.uchar)(unsafe.Pointer(&ret[0]))

	C.ECC_point_compress(pin, C.ushort(len(point)), pout, C.uint(typeChoose))
	return ret[:]
}

func PointDecompress(point []byte, typeChoose uint32) []byte {
	ret := [65]byte{}
	pin := (*C.uchar)(unsafe.Pointer(&point[0]))
	pout := (*C.uchar)(unsafe.Pointer(&ret[0]))

	C.ECC_point_decompress(pin, C.ushort(len(point)), pout, C.uint(typeChoose))

	return ret[:]
}

func Hash(data []byte, digestLen uint16, typeChoose uint32) []byte {
	var length uint16
	switch typeChoose {
	case HASH_ALG_MD4, HASH_ALG_MD5:
		length = 16
		break
	case HASH_ALG_SHA1, HASH_ALG_RIPEMD160, HASH_ALG_HASH160:
		length = 20
		break
	case HASH_ALG_SHA256, HASh_ALG_DOUBLE_SHA256, HASH_ALG_SM3:
		length = 32
		break
	case HASH_ALG_SHA512:
		length = 64
		break
	case HASH_ALG_BLAKE2B, HASH_ALG_BLAKE2S:
		length = digestLen
		break
	default:
		break
	}
	ret := make([]byte, length)

	msg := (*C.uchar)(unsafe.Pointer(&data[0]))
	dig := (*C.uchar)(unsafe.Pointer(&ret[0]))

	C.hash(msg, C.uint(len(data)), dig, C.ushort(digestLen), C.uint(typeChoose))

	return ret[:]
}
