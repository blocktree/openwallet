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

#ifndef hmac_h
#define hmac_h
#include "stdio.h"
#include "string.h"
#include "stdlib.h"
#include "sha256.h"
#include "sha512.h"
#include "sm3.h"
#include "type.h"

#define  HMAC_SHA256_ALG        0x50505050
#define  HMAC_SHA512_ALG        0x50505051
#define  HMAC_SM3_ALG           0x50505052


/*
 @function:compute massage authentication code
 @paramter[in]:K pointer to key
 @paramter[in]:Klen,the byte length of K
 @paramter[in]:M pointer to message to be authenticated
 @paramter[in]:Mlen,the byte length of M
 @paramter[out]:out pointer to HMAC result
 @paramter[in]:Hmac_Hash_Alg,hash algorithm flag.
 HAMC_SHA256_ALG: SHA256
 HMAC_SHA512_ALG: SHA512
 HMAC_SM3_ALG:SM3
 default:not support.
 @return:NULL
 @notice:if HAMC_SHA256_ALG,out space is 32 byte,if HMAC_SHA512_ALG,out space is 64 byte; if HMAC_SM3_ALG,out space is 32 byte
 */
void HMAC(uint8_t *K,uint16_t Klen,uint8_t *M,uint16_t Mlen,uint8_t *out,uint32_t Hmac_Hash_Alg);
#endif /* end */
