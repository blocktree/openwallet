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

#ifndef _MD5_H__
#define _MD5_H__
#include <stdio.h>
#include <string.h>
#include "type.h"
/* MD5 context. */
typedef struct {
    uint32_t state[4];                /* intermediate state (ABCD) */
    uint32_t count[2];                /* number of bits, modulo 2^64 */
    uint8_t buffer[64];               /* input buffer */
} MD5_CTX;

/*
 @function:init MD5_CTX,writing a new message
 @paramter[in]:ctx pointer to MD5_CTX
 @return: NULL
 @notice: none
 */
void  md5_init (MD5_CTX *ctx);
/*
 @function:update message Continues an MD5 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to MD5_CTX
 @paramter[in]:msg pointer to the message to do md5
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void md5_update(MD5_CTX * ctx, const uint8_t *msg, const uint32_t msg_len);
/*
 @function:finalization md5 operation Ends an MD5 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to MD5_CTX
 @paramter[out]:digest pointer to md5 hash result
 @return:NULL
 @notice:nothing
 */
void  md5_final (MD5_CTX *ctx,uint8_t digest[16]);

/*
 @function:md5 hash
 @paramter[in]:msg pointer to the message to do md5
 @paramter[in]:msg_len,the byte length of input
 @digest[out]:digest pointer to md5 hash result
 @return:NULL
 @notice:none
 */
void  md5_hash(const uint8_t *msg,uint32_t msg_len,uint8_t digest[16]);

#endif    /* _MD5_H__ */


