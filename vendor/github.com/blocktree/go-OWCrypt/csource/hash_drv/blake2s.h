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

#ifndef blake2s_h
#define blake2s_h

#include <stdio.h>
#include <string.h>
#include <assert.h>
#include "type.h"
#define BLAKE2S_DIGEST_LENGTH 32
#define BLAKE2S_BLOCKBYTES    64
#define BLAKE2S_OUTBYTES      32
#define BLAKE2S_KEYBYTES      32
#define BLAKE2S_SALTBYTES     8
#define BLAKE2S_PERSONALBYTES 8

struct blake2s_param_st
{
    uint8_t  digest_length; //digest byte length, an integer in [1,32]
    uint8_t  key_length;    //key byte length, an integer in [0,32]
    uint8_t  fanout;        //(1 byte)an integer in[0,255](set to 0 if unlimited,and to 1 only in sequential mode)
    uint8_t  depth;         //(1 byte) an integer in [1,255](set to 255 if unlimited,and to 1 only in sequential mode)
    uint8_t  leaf_length[4];//(4 bytes)an integer in[0,2^32-1]
    uint8_t  node_offset[6];// 6 bytes,range in[1,2^48-1](set to 0 for the first, leftmost, leaf, or in sequential mode)
    uint8_t  node_depth;    //1 byte,an integer in [1, 255] (set to 255 if unlimited, and to 1 only in sequential mode)
    uint8_t  inner_length;  //1 byte,range in [0,32](set to 0 in sequential mode)
    uint8_t  salt[BLAKE2S_SALTBYTES]; //8 bytes,(set to all-NULL by default)
    uint8_t  personal[BLAKE2S_PERSONALBYTES];  //8 bytes,(set to all-NULL by default)
};
struct blake2s_ctx_st
{
    uint32_t h[8]; //store hash intermidiate state
    uint32_t t[2];//counter
    uint32_t f[2];//finalization flags
    uint8_t  buf[BLAKE2S_BLOCKBYTES];//store block message to deal with
    size_t   buflen; //the message byte length left
};

typedef struct blake2s_param_st BLAKE2S_PARAM;
typedef struct blake2s_ctx_st BLAKE2S_CTX;

/*
 @function:init BLAKE2S_CTX,writing a new message
 @paramter[in]:ctx pointer to BLAKE2S_CTX structure
 @paramter[in]:key pointer to the key(if dosen't need key,please input NULL)
 @paramter[in]:key_len denotes the byte length of key.(if dosen't need key,please set key_bytelen to zero)
 @paramter[in]:digest_len denotes the expected hash result length
 */
void blake2s_init(BLAKE2S_CTX *ctx,uint8_t *key,uint16_t key_len,uint16_t digest_len);
/*
 @function:update message Continues an blake2b message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:c pointer to BLAKE2S_CTX structure
 @paramter[in]:data pointer to the message to do hash
 @paramter[in]:datalen denotes the byte length of data.
 */
void blake2s_update(BLAKE2S_CTX *ctx, const uint8_t *msg, uint32_t mas_len);

/*
 @function: end an ripemd160 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to BLAKE2S_CTX structure
 @paramter[out]:digest pointer to hash intermidate intermidiate result
 @paramter[in]:digest_len denotes the byte length of digest
 */
void blake2s_final(BLAKE2S_CTX *ctx,uint8_t *digest,uint32_t digest_len);

/*
 @function:BLAKE2s hash
 @paramter[in]:msg pointer to the data to do hash
 @paramter[in]:msg_len denotes the byte length of msg
 @paramter[in]:key pointer to the key(if dosen't need key,please input NULL)
 @paramter[in]:key_len denotes the byte length of key.(if dosen't need key,please set key_bytelen to zero)
 @paramter[in]:digest_len denotes the expected hash result length(rang in[1,32])
 @paramter[out]:digest pointer to hash result
 */
void blake2s(uint8_t *msg, uint16_t msg_len,uint8_t *key,uint16_t key_length, uint8_t digest_len, uint8_t *digest);

#endif /* blake2s_h */
