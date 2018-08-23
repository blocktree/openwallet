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

#ifndef secp256r1_h
#define secp256r1_h

#include <stdio.h>
#include "ECDSA.h"
#include "ecc_set.h"
#include "type.h"

void secp256r1_get_order(uint8_t *order);

uint16_t secp256r1_genPubkey(uint8_t *prikey, uint8_t *pubkey);
uint16_t secp256r1_sign(uint8_t *prikey, uint8_t *message, uint16_t message_len, uint8_t *sig);
uint16_t secp256r1_verify(uint8_t *pubkey, uint8_t *message, uint16_t message_len, uint8_t *sig);
/*
 @function:(Point) outpoint_buf= (Point)inputpoint1_buf+[k](Point)inputpoint2_buf
 @paramter[in]:inputpoint1_buf pointer to one point (stored by byte string)on the curve elliptic
 @paramter[in]:Q pointer to another point(stored by byte string) on the curve elliptic
 @paramter[in]:k pointer to the multiplicator
 @paramter[in]:outpoint_buf pointer to the result(stored by byte string)
 @return:0表示运算失败；1表示运算成功.
 */
uint16_t secp256r1_point_mul_add(uint8_t *inputpoint1_buf,uint8_t *inputpoint2_buf,uint8_t *k,uint8_t *outpoint_buf);

/*
 @function:(Point)outpoint_buf = (Point)inputpoint_buf+[k]G(G is the base point of curve elliptic)
 @paramter[in]:inputpoint_buf pointer to the point on curve elliptic(stored by byte string)
 @paramter[in]:k pointer to the multiplicator
 @paramter[out]:outpoint_buf pointer to the result(stored by byte string)
 @return:0 表示运算失败；1表示运算成功.
 */

uint16_t secp256r1_point_mul_base_G_add(uint8_t *inputpoint_buf,uint8_t *k,uint8_t *outpoint_buf);

/*
 @function:椭圆曲线（参数为secp256r1）上点的压缩
 @paramter[in]:point_buf,待压缩的点(stored by byte string)
 @paramter[in]:point_buf_len表示point_buf的字节长度
 @paramter[out]:x,点压缩后的横坐标（长度为ECC_LEN+1 字节）
 @return：0表示压缩失败；1表示压缩成功
 */

uint16_t secp256r1_point_compress(uint8_t *point_buf,uint16_t point_buf_len,uint8_t *x);

/*
 @function:椭圆曲线(参数为secp256r1)点的解压缩
 @paramter[in]:x pointer to the x-coordiate of the point on curve elliptic
 @paramter[in]:x_len denotes the byte length of x(x_len=ECC_LEN=1)
 @paramter[in]:TYpe denotes ECC_CURVE_PARAM type
 @paramter[out]:point_buf pointer to the xy-coordiate(with 0x04) of the point on curve elliptic
 @return：0 表示解压缩失败；1 表示解压缩成功.
 */
uint16_t secp256r1_point_decompress(uint8_t *x,uint16_t x_len,uint8_t *point_buf);
#endif /* secp256r1_h */
