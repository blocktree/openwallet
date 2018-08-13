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
#include "sha1.h"
#include "sha256.h"
#include "sha512.h"
#include "sm3.h"
#include "md5.h"
#include "ripemd160.h"
#include "blake2b.h"
#include "blake2s.h"
#include "type.h"
#ifndef MD5_ALG
#define MD5_ALG    0
#endif

#ifndef SHA1_ALG
#define SHA1_ALG   1
#endif

#ifndef SHA256_ALG
#define SHA256_ALG 2
#endif

#ifndef SHA512_ALG
#define SHA512_ALG 3
#endif

#ifndef SM3_ALG
#define SM3_ALG    4
#endif

#ifndef RIPEMD160
#define RIPEMD160 5
#endif

#ifndef BLAKE2B
#define BLAKE2B   6
#endif

#ifndef BLAKE2S
#define BLAKE2S   7
#endif

/*
 @function:compute massage authentication code
 @paramter[in]:K pointer to key
 @paramter[in]:Klen,the byte length of K
 @paramter[in]:M pointer to message to be authenticated
 @paramter[in]:Mlen,the byte length of M
 @paramter[out]:out pointer to HMAC result
 @paramter[in]:Hash_Alg,hash algorithm flag.if Hash_Alg = MD5_ALG,Choose MD5 algorithm;if Hash_Alg = SHA1_ALG,choose SHA1 algorithm;
 Hash_Alg=SHA256_ALG,Choose SHA256 algorithm;if Hash_Alg=SHA512_ALG, Choose SHA512 algorithm;if Hash_Alg=SM3_ALG, Choose SM3 algorithm;if HAsh_Alg=BLAKE2B_ALG, choose BLAKE2B algorithm;if HAsh_Alg=BLAKE2S,choose BLAKE2S algorithm;default:not support.
 @return:NULL
 @notice:if Hash_Alg=MD5_ALG,the space size of out is 16 byte; if Hash_Alg=SHA1_ALG,the space size of out is 20 byte; if Hash_Alg=SHA256_ALG,the space size of out is 32 byte;if Hash_Alg=SHA512_ALG,the space size of out is 64 byte; if Hash_Alg=SM3_ALG,the space size of out is 32 byte;
 */
void HMAC(uint8_t *K,uint16_t Klen,uint8_t *M,uint16_t Mlen,uint8_t *out,uint8_t Hash_Alg);
#endif /* end */
