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

#ifndef blake256_h
#define blake256_h

#include "type.h"
#include "string.h"


typedef struct
{
    uint32_t h[8], s[4], t[2];
    int buflen, nullt;
    uint8_t  buf[64];
} BLAKE256_CTX;


/*
 @function:init BLAKE256_CTX,writing a new message
 @paramter[in]:ctx pointer to BLAKE256_CTX
 @return: NULL
 @notice: none
 */
void blake256_init( BLAKE256_CTX *ctx);

/*
 @function:update message Continues an blake256 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to BLAKE256_CTX
 @paramter[in]:msg pointer to the message to do blake256
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void blake256_update( BLAKE256_CTX *ctx, const uint8_t *msg, uint64_t msg_len );

/*
 @function:finalization blake256 operation ends an blake256 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to BLAKE256_CTX
 @paramter[out]:digest pointer to blake256 hash result
 @return:NULL
 @notice:nothing
 */
void blake256_final( BLAKE256_CTX *ctx, uint8_t *digest);

/*
 @function: blake256 hash
 @parameter[in]:msg pointer to the message to do hash
 @parameter[in]:msg_len,the byte length of msg
 @parameter[in]:digest pointer to hash result
 @return: none
 @notice:nothing
 */
void blake256_hash(const uint8_t *msg, uint64_t msg_len,uint8_t *digest);


#endif /* blake_h */
