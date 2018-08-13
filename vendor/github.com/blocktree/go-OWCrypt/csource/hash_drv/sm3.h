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

#ifndef sm3_h
#define sm3_h

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "type.h"
#ifndef SM3_LBLOCK
#define SM3_LBLOCK           16
#endif

#ifndef SM3_CBLOCK
#define SM3_CBLOCK           (SM3_LBLOCK * 4)
#endif

#ifndef SM3_DIGEST_LENGTH
#define SM3_DIGEST_LENGTH    32
#endif

typedef struct SM3_state_st
{
    uint32_t h0,h1,h2,h3,h4,h5,h6,h7;  /*intermediate state*/
    uint32_t Nl,Nh;   /* number of bits, modulo 2^64 */
    uint32_t data[68]; /*input buffer*/
} SM3_CTX;

/*
 @function:init SM3_CTX,writing a new message
 @paramter[in]:ctx pointer to SM3_CTX
 @return: NULL
 @notice: none
 */
void sm3_init(SM3_CTX *ctx);

/*
 @function:update message Continues an sm3 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to SM3_CTX
 @paramter[in]:msg pointer to the message to do sm3
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void sm3_update(SM3_CTX *ctx, uint8_t *msg, size_t msg_len);

/*
 @function:finalization sm3 operation ends an sha1 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to SM3_CTX
 @paramter[out]:digest pointer to sm3 hash result
 @return:NULL
 @notice:nothing
 */
void sm3_final(SM3_CTX *ctx, uint8_t *digest);

/*
 @function: sm3 hash
 @parameter[in]:msg pointer to the message to do hash
 @parameter[in]:msg_len,the byte length of msg
 @parameter[in]:digest pointer to hash result
 @return: none
 @notice:nothing
 */
void sm3_hash(uint8_t *msg, uint32_t msg_len, uint8_t *hash);

#endif /* sm3_h */
