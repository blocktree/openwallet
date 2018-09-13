/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

#ifndef ecc_set_h
#define ecc_set_h

#include <stdio.h>
#include "ecc_drv.h"
#include "type.h"


#define  ECC_CURVE_SECP256K1        0xECC00000
#define  ECC_CURVE_SECP256R1        0xECC00001
#define  ECC_CURVE_PRIMEV1          ECC_CURVE_SECP256R1
#define  ECC_CURVE_NIST_P256        ECC_CURVE_SECP256R1
#define  ECC_CURVE_SM2_STANDARD     0xECC00002
#define  ECC_CURVE_ED25519          0xECC00003
#define  ECC_CURVE_ED25519_EXTEND   0xECC00004

#define SUCCESS              0x0001
#define FAILURE              0x0000
#define ECC_PRIKEY_ILLEGAL   0xE000
#define ECC_PUBKEY_ILLEGAL   0xE001
#define ECC_WRONG_TYPE       0xE002
#define ECC_MISS_ID          0xE003

uint16_t ECC_genPubkey(uint8_t *prikey, uint8_t *pubkey, uint32_t type);
uint16_t ECC_sign(uint8_t *prikey, uint8_t *ID, uint16_t IDlen, uint8_t *message, uint16_t message_len, uint8_t *sig, uint32_t type);
uint16_t ECC_verify(uint8_t *pubkey, uint8_t *ID, uint16_t IDlen, uint8_t *message, uint16_t message_len, uint8_t *sig, uint32_t type);
uint16_t ECC_enc(uint8_t *pubkey, uint8_t *plain, uint16_t plain_len, uint8_t *cipher, uint16_t *cipher_len, uint32_t type);
uint16_t ECC_dec(uint8_t *prikey, uint8_t *cipher, uint16_t cipher_len, uint8_t *plain, uint16_t *plain_len, uint32_t type);

//////////////////////////////////////////////////////协商/////////////////////////////////////////////////
uint16_t ECC_key_exchange_initiator_step1(uint8_t *tmpPriInitiator, uint8_t *tmpPubInitiator, uint32_t type);
uint16_t ECC_key_exchange_initiator_step2(uint8_t *IDinitiator,         \
                                          uint16_t IDinitiator_len,     \
                                          uint8_t *IDresponder,         \
                                          uint16_t IDresponder_len,     \
                                          uint8_t *priInitiator,        \
                                          uint8_t *pubInitiator,        \
                                          uint8_t *pubResponder,        \
                                          uint8_t *tmpPriInitiator,     \
                                          uint8_t *tmpPubInitiator,     \
                                          uint8_t *tmpPubResponder,     \
                                          uint8_t *Sin,                 \
                                          uint8_t *Sout,                \
                                          uint16_t keylen,              \
                                          uint8_t *key,                 \
                                          uint32_t type);
uint16_t ECC_key_exchange_responder_step1(uint8_t *IDinitiator,         \
                                          uint16_t IDinitiator_len,     \
                                          uint8_t *IDresponder,         \
                                          uint16_t IDresponder_len,     \
                                          uint8_t *priResponder,        \
                                          uint8_t *pubResponder,        \
                                          uint8_t *pubInitiator,        \
                                          uint8_t *tmpPubResponder,     \
                                          uint8_t *tmpPubInitiator,     \
                                          uint8_t *Sin,                 \
                                          uint8_t *Sout,                \
                                          uint16_t keylen,              \
                                          uint8_t *key,                 \
                                          uint32_t type);
uint16_t ECC_key_exchange_responder_step2(uint8_t *Sinitiator, uint8_t *Sresponder, uint32_t type);
/*
 @function:(Point)outpoint_buf = (Point)inputpoint1_buf+[k](Point)inputpoint2_buf
 @paramter[in]:inputpoint1_buf pointer to one point(stored by byte string) on the curve elliptic
 @paramter[in]:inputpoint2_buf pointer to another point(stored by byte string) on the curve elliptic
 @paramter[in]:k pointer to the multiplicator
 @paramter[out]:outpoint_buf pointer to the result(stored by byte string)
 @paramter[in]:type denotes ECC_CURVE_PARAM type.ECC_CURVE_SECP256K1:choose secp256k1 paramters;ECC_CURVE_SECP256R1:choose
 secp256r1 paramters; ECC_CURVE_SM2_STANDARD;choose sm2 paramters.others:not support.
 @return:0表示运算失败；1表示运算成功.
 */

//uint16_t ECC_point_mul_add(ECC_POINT *P,ECC_POINT *Q,uint8_t *k,ECC_POINT *T,uint32_t Type)
uint16_t ECC_point_mul_add(uint8_t *inputpoint1_buf,uint8_t *inputpoint2_buf,uint8_t *k,uint8_t *outpoint_buf,uint32_t type);

/*
 @function:(Point)outpoint_buf = (Point)inputpoint_buf+[k]G(G is the base point of curve elliptic)
 @paramter[in]:inputpoint_buf pointer to one point(stored by byte string) on the curve elliptic
 @paramter[in]:k pointer to the multiplicator
 @paramter[out]:outpoint_buf pointer to the result(stored by byte string)
 @paramter[in]:type denotes ECC_CURVE_PARAM type.ECC_CURVE_SECP256K1:choose secp256k1 paramters;ECC_CURVE_SECP256R1:choose
 secp256r1 paramters; ECC_CURVE_SM2_STANDARD;choose sm2 paramters.others:not support.
 @return:0 表示运算失败；1 表示运算成功.
 */

uint16_t ECC_point_mul_baseG_add(uint8_t *inputpoint_buf,uint8_t *k,uint8_t *outpoint_buf,uint32_t type);

uint16_t ECC_point_mul_baseG(uint8_t *scalar, uint8_t *point, uint32_t type);

/*
 @function:椭圆曲线上点的压缩
 @paramter[in]:pubKey,待压缩的公钥
 @paramter[in]:pubKey_len表示公钥的字节长度
 @paramter[out]:x,公钥压缩后的横坐标（长度为ECC_LEN+1 字节）
 @paramter[in]:TYpe denotes ECC_CURVE_PARAM type.ECC_CURVE_SECP256K1:choose secp256k1 paramters;ECC_CURVE_SECP256R1:choose
 secp256r1 paramters; ECC_CURVE_SM2_STANDARD;choose sm2 paramters.others:not support.
 @return：0 表示压缩失败；1 表示压缩成功
 @note:secp256k1/secp256r1/sm2三种形式的参数，点的压缩都是一样的处理流程.此处之所以通过Type做区别，只是为了在形式上与解压缩函数保持一致.
 */

uint16_t ECC_point_compress(uint8_t *pubKey,uint16_t pubKey_len,uint8_t *x,uint32_t type);


/*
 @function:椭圆曲线上点的解压缩
 @paramter[in]:curveParam pointer to curve elliptic paramters
 @paramter[in]:x pointer to the x-coordiate of the point on curve elliptic
 @paramter[in]:x_len denotes the byte length of x(x_len=ECC_LEN=1)
 @paramter[out]:y pointer to the y-coordiate of the point on curve elliptic
 @paramter[in]:Type denotes ECC_CURVE_PARAM type.ECC_CURVE_SECP256K1:choose secp256k1 paramters;ECC_CURVE_SECP256R1:choose
 secp256r1 paramters; ECC_CURVE_SM2_STANDARD;choose sm2 paramters.others:not support.
 @return:0 表示解压缩失败；1 表示解压缩成功
 */
uint16_t ECC_point_decompress(uint8_t *x,uint16_t x_len,uint8_t *y,uint32_t type);

/*
 @function:获取椭圆曲线的阶
 @paramter[in]:Type denotes ECC_CURVE_PARAM type.ECC_CURVE_SECP256K1:choose secp256k1 paramters;ECC_CURVE_SECP256R1:choose
 @paramter[out]:order the order of the curve
 @return: SUCCESS/ECC_WRONG_TYPE
 */
uint16_t ECC_get_curve_order(uint8_t *order, uint32_t type);

#endif /* ecc_set_h */
