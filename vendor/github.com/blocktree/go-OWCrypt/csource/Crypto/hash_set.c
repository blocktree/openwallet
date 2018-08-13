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

#include "hash_set.h"
/*
 @function:hash operation
 @paramter[in]:msg pointer to the message to do hash
 @paramter[in]:msg_len denotes the byte length of msg
 @paramter[in]:type,hash algorithm flag.if type = HASH_ALG_SHA1,choose sha1 algorithm; if type=HASH_ALG_SHA256,choose sha256 algorithm; if type=HASH_ALG_SHA512,
 choose sha512 algorithm; if type=HASH_ALG_SM3,choose sm3 algorithm;if Type=HASH_ALG_MD5, choose md5 algorithm; if Type=HASH_ALG_RIPEMD160,choose ripemd160 algorithm;if type=HASH_ALG_BLAKE2B,choose blake2b algorithm; if type = HASH_ALG_BLAKE2S, choose blake2s algorithm.if type=HASh_ALG_DOUBLE_SHA256,choose two consecutive sha256;if type =HASH_ALG_SHA256_RIPEMD160, first do sha256,then do ripemed160. otherwise,not support.
 @paramter[out]:digest pointer to hash result(make sure the space size is enough)
 @paramter[in]:digest_len,the byte length of digest.It is useful if and only if blake2b and blake2s algorithm.Because the digest length of other hash algorithms is fix.
 */
void hash(uint8_t *msg,uint32_t msg_len,uint8_t *digest,uint16_t digest_len,uint32_t type)
{
    switch (type)
    {
        case HASH_ALG_SHA1:
            sha1_hash(msg, msg_len,  digest);
            break;
        case HASH_ALG_SHA256:
            sha256_hash(msg, msg_len, digest);
            break;
        case HASH_ALG_SHA512:
            sha512_hash(msg, msg_len, digest);
            break;
        case HASH_ALG_SM3:
            sm3_hash(msg, msg_len, digest);
            break;
        case HASH_ALG_MD4:
            md4_hash(msg,msg_len,digest);
            break;
        case HASH_ALG_MD5:
            md5_hash(msg,msg_len,digest);
            break;
        case HASH_ALG_RIPEMD160:
            ripemd160_hash(msg,msg_len,digest);
            break;
        case HASH_ALG_BLAKE2B:
            blake2b(msg, msg_len,NULL,0, digest_len, digest);
            break;
        case HASH_ALG_BLAKE2S:
            blake2s(msg, msg_len,NULL,0, digest_len, digest);
            break;
        case HASh_ALG_DOUBLE_SHA256:
            sha256_hash(msg, msg_len, digest);
            sha256_hash(digest, 32, digest);
            break;
        case HASH_ALG_SHA256_RIPEMD160:
            sha256_hash(msg, msg_len, digest);
            ripemd160_hash(digest,32,digest);
            break;
        default:
            break;
    }
}
