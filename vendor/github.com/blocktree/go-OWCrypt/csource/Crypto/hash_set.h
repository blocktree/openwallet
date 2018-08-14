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

#ifndef hash_set_h
#define hash_set_h

#include <stdio.h>
#include "sha1.h"
#include "sha256.h"
#include "sha512.h"
#include "sm3.h"
#include "md5.h"
#include "ripemd160.h"
#include "blake2b.h"
#include "blake2s.h"
#include "md4.h"
#include "type.h"

#define HASH_ALG_SHA1              0xA0000000
#define HASH_ALG_SHA256            0xA0000002
#define HASH_ALG_SHA512            0xA0000003
#define HASH_ALG_MD4               0xA0000004
#define HASH_ALG_MD5               0xA0000005
#define HASH_ALG_RIPEMD160         0xA0000006
#define HASH_ALG_BLAKE2B           0xA0000007
#define HASH_ALG_BLAKE2S           0xA0000008
#define HASH_ALG_SM3               0xA0000009
#define HASh_ALG_DOUBLE_SHA256     0xA000000A
#define HASH_ALG_SHA256_RIPEMD160  0xA000000B

/*
 @function:hash operation
 @paramter[in]:msg pointer to the message to do hash
 @paramter[in]:msg_len denotes the byte length of msg
 @paramter[in]:type,hash algorithm flag.if type = HASH_ALG_SHA1,choose sha1 algorithm; if type=HASH_ALG_SHA256,choose sha256 algorithm; if type=HASH_ALG_SHA512,
 choose sha512 algorithm; if type=HASH_ALG_SM3,choose sm3 algorithm;if Type=HASH_ALG_MD5, choose md5 algorithm; if Type=HASH_ALG_RIPEMD160,choose ripemd160 algorithm;if type=HASH_ALG_BLAKE2B,choose blake2b algorithm; if type = HASH_ALG_BLAKE2S, choose blake2s algorithm.if type=HASh_ALG_DOUBLE_SHA256,choose two consecutive sha256;if type =HASH_ALG_SHA256_RIPEMD160, first do sha256,then do ripemed160. otherwise,not support.
 @paramter[out]:digest pointer to hash result(make sure the space size is enough)
 @paramter[in]:digest_len,the byte length of digest.It is useful if and only if blake2b and blake2s algorithm.Because the digest length of other hash algorithms is fix.
 */
void hash(uint8_t *msg,uint32_t msg_len,uint8_t *digest,uint16_t digest_len,uint32_t Type);

#endif /* hash_set_h */
