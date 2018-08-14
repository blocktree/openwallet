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

#ifndef ripemd160_h
#define ripemd160_h

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "type.h"
typedef struct {
    uint32_t total[2];  /* number of bits, modulo 2^64 */
    uint32_t state[5];  /*intermediate state*/
    uint32_t buffer[16]; /*input buffer*/
}RIPEMD160_CTX;

/*
 @function:init RIPEMD160_CTX,writing a new message
 @paramter[in]:ctx pointer to RIPEMD160_CTX
 @return: NULL
 @notoce: none
 */
void ripemd160_init(RIPEMD160_CTX *ctx);

/*
 @function:update message Continues an ripemd160 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to RIPEMD160_CTX
 @paramter[in]:msg pointer to the message to do ripemd160
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void ripemd160_update(RIPEMD160_CTX *ctx,uint8_t *msg,uint32_t msg_len);

/*
 @function: end an ripemd160 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to RIPEMD160_CTX
 @paramter[out]:digest pointer to md5 hash result
 @return:NULL
 @notice:nothing
 */
void ripemd160_final(RIPEMD160_CTX *ctx,uint8_t digest[20]);

/*
 @function:ripemd160 hash
 @paramter[in]:msg pointer to the message to do ripemd160
 @paramter[in]:msg_len,the byte length of input
 @digest[out]:digest piointer to  ripemd160 hash result
 @return:NULL
 @notice:none
 */
void ripemd160_hash(uint8_t *msg,uint32_t msg_len,uint8_t digest[20]);
#endif /* ripemd160_h */
